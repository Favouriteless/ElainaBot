package discord

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const defaultQueueSize = 16

// Websocket disconnect codes as specified by https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-close-event-codes
// Codes not needed for responses have not been included (e.g. encoding errors)
const (
	closeUnknownError         = 4000
	closeNotAuthenticated     = 4003
	closeAuthenticationFailed = 4004
	closeInvalidSequence      = 4007
	closeRateLimited          = 4008
	closeTimedOut             = 4009
)

var gatewayConnection *gateway

var ErrInvalidResumeUrl = errors.New("invalid resume url")

type gateway struct {
	conn      *websocket.Conn // Websocket connection for the client. Only supports writing from one thread, use sendQueue to write to it
	sendQueue chan []byte     // Buffer for writing to conn, can still accept items even when disconnected

	cardiacArrest atomic.Bool    // cardiacArrest will stop all gateway child goroutines (handleWriting, heartbeat, readUntilClosed)
	wg            sync.WaitGroup // wg will allow disconnect to block until all the child processes are finished.

	intents int // gateway intents the bot will be using https://discord.com/developers/docs/events/gateway#gateway-intents

	connectUrl string
	resuming   bool // If true, gateway.connect will attempt to resuming the previous session
	resumeUrl  string
	sessionId  string       // ID of gateway session, only applicable if resuming
	sequence   atomic.Int32 // The last sequence number the client received from the gateway

	heartbeatAcknowledged bool // Update to false when the client sends a sendHeartbeat. If discord doesn't acknowledge before the next sendHeartbeat, reconnect.
}

// ListenGateway initializes the gateway struct and attempts to send a connection request to the API.
func ListenGateway(intents int, quit <-chan interface{}) {
	slog.Info("Initializing gateway connection...")

	go func() {
		gatewayConnection = &gateway{sendQueue: make(chan []byte, defaultQueueSize), intents: intents}

		go func() {
			<-quit
			slog.Info("Closing gateway connection")
			gatewayConnection.cardiacArrest.Store(true)
			_ = gatewayConnection.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		}()

		for {
			reconnect, err := gatewayConnection.connect()
			if err != nil {
				slog.Error("Gateway connection error: " + err.Error())
			}
			if !reconnect {
				break
			}
		}
	}()
}

// connect attempts to initialize a gateway connection. If resuming is true, the connection will attempt to resuming the
// last session.
func (gateway *gateway) connect() (reconnect bool, err error) {
	url, err := gateway.getConnectUrl()
	if errors.Is(err, ErrInvalidResumeUrl) {
		gateway.resuming = false
		return gateway.connect()
	} else if err != nil {
		return false, err
	}

	slog.Info("Attempting to connect to gateway:", slog.String("api_version", apiVersion), slog.String("api_encoding", apiEncoding))
	gateway.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		if os.IsTimeout(err) {
			gateway.connectUrl = ""
			return true, err
		}
		return false, err
	}
	slog.Info("Websocket initialized")

	gateway.cardiacArrest.Store(false)

	var closeCode int
	gateway.conn.SetCloseHandler(func(c int, t string) error {
		closeCode = c
		return nil
	}) // Capture the close code after the wg finishes

	go gateway.handleWriting()
	gateway.resuming, err = gateway.readUntilClosed(gateway.resuming)

	gateway.cardiacArrest.Store(true) // Stop children & reset the connection state after the websocket disconnects
	gateway.wg.Wait()

	if err == nil { // No error means the reader received a reconnect request
		return true, nil
	}

	if closeCode == closeUnknownError || closeCode == closeNotAuthenticated || closeCode == closeInvalidSequence ||
		closeCode == closeRateLimited || closeCode == closeTimedOut { // If not one of these, we should be safe and not attempt to resume
		gateway.resuming = true
	}

	return true, err
}

// readUntilClosed reads all payloads from a websocket connection and dispatches them to the event handler.
// No stop flag is needed as the connection will return an error when it closes anyway.
func (gateway *gateway) readUntilClosed(resuming bool) (shouldResume bool, err error) {
	for {
		if gateway.cardiacArrest.Load() {
			return false, nil
		}
		var payload gatewayPayload
		if err := gateway.conn.ReadJSON(&payload); err != nil {
			return false, err
		}

		switch payload.Opcode {
		case opHello:
			var hello HelloPayload
			if err = json.Unmarshal(*payload.Data, &hello); err != nil {
				panic(err) // Should never be hit
			}
			slog.Info("Starting heartbeats")
			go gateway.heartbeat(time.Millisecond * time.Duration(hello.HeartbeatInterval))
			gateway.identify(resuming)
		case opDispatch:
			gateway.sequence.Store(*payload.SequenceNum)
			go dispatchEvent(*payload.EventName, *payload.Data)
		case opHeartbeat:
			gateway.sendHeartbeat()
		case opReconnect:
			slog.Info("Discord requested a gateway reconnect")
			return true, nil
		case opInvalidSession:
			slog.Warn("Discord invalidated the gateway session. Attempting to reconnect...")
			var resume bool
			if err = json.Unmarshal(*payload.Data, &resume); err != nil {
				panic(err) // Should never be hit
			}
			return resume, nil
		case opHeartbeatAck:
			gateway.heartbeatAcknowledged = true
		}
	}
}

// gatewayPayload represents a gateway payload for discord API. Pointers are for optional fields to encode as NULL
// See: https://discord.com/developers/docs/events/gateway-events#payload-structure
type gatewayPayload struct {
	Opcode      int              `json:"op"`
	Data        *json.RawMessage `json:"d"`
	SequenceNum *int32           `json:"s,omitempty"`
	EventName   *string          `json:"t,omitempty"`
}

// sendPayload sends the given gatewayPayload on the websocket connection if one is open, or errors if there is no valid
// connection or the payload fails to encode
func (gateway *gateway) sendPayload(payload *gatewayPayload) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	gateway.sendQueue <- encoded
	return nil
}

// Identify sends either an IDENTIFY or RESUME packet depending on the value of resuming
func (gateway *gateway) identify(resume bool) {
	if resume {
		id, err := json.Marshal(ResumePayload{
			Token:       application.token,
			SessionId:   gateway.sessionId,
			SequenceNum: gateway.sequence.Load(),
		})
		if err != nil {
			panic(err) // Should never be hit
		}
		if err = gateway.sendPayload(&gatewayPayload{Opcode: opResume, Data: (*json.RawMessage)(&id)}); err != nil {
			panic(err) // Should never be hit
		}
		slog.Info("Resuming gateway connection")
		return
	}
	enc, err := json.Marshal(IdentifyPayload{
		Token:      application.token,
		Properties: ConnectionProperties{Os: "windows", Browser: application.name, Device: application.name},
		Intents:    gateway.intents,
	})
	if err != nil {
		panic(err) // Should never be hit
	}
	if err = gateway.sendPayload(&gatewayPayload{Opcode: opIdentify, Data: (*json.RawMessage)(&enc)}); err != nil {
		panic(err) // Should never be hit
	}
	slog.Info("Identifying gateway connection")
}

func (gateway *gateway) sendHeartbeat() {
	data, err := json.Marshal(gateway.sequence.Load())
	if err != nil {
		panic(err) // should never be hit
	}

	if err = gateway.sendPayload(&gatewayPayload{Opcode: opHeartbeat, Data: (*json.RawMessage)(&data)}); err != nil {
		slog.Error("Failed to send heartbeat: " + err.Error())
		return
	}
	gateway.heartbeatAcknowledged = false
}

func (gateway *gateway) heartbeat(interval time.Duration) {
	gateway.wg.Add(1)
	defer gateway.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	gateway.sendHeartbeat()
	for {
		select {
		case <-ticker.C:
			if !gateway.heartbeatAcknowledged {
				slog.Error("Heartbeat not acknowledged, terminating connection to resume")
				gateway.resuming = true
				_ = gateway.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseServiceRestart, ""), time.Now().Add(time.Second))
				gateway.conn.Close()
				return
			}

			gateway.sendHeartbeat()
		default:
			if gateway.cardiacArrest.Load() {
				return
			}
		}
	}
}

func (gateway *gateway) handleWriting() {
	gateway.wg.Add(1)
	defer gateway.wg.Done()
	for {
		select {
		case payload := <-gateway.sendQueue:
			if err := gateway.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				slog.Error("Failed write to gateway connection: " + err.Error()) // This should never be hit, but just in case
			}
		default:
			if gateway.cardiacArrest.Load() {
				return
			}
		}
	}
}

func (gateway *gateway) getConnectUrl() (string, error) {
	if gateway.resuming {
		if gateway.resumeUrl == "" {
			return "", ErrInvalidResumeUrl
		}
		return gateway.resumeUrl, nil
	}
	if gateway.connectUrl != "" {
		return gateway.connectUrl, nil
	}
	return fetchGatewayUrl()
}

func fetchGatewayUrl() (string, error) {
	resp, err := Get(Url("gateway/bot"))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err) // Should never be hit
		}
		return "", fmt.Errorf("failed to fetch gateway url: %s: %s", resp.Status, string(body))
	}

	var data struct{ Url string }
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.Url == "" {
		return "", errors.New("could not fetch gateway url. this is probably an issue on discord's end")
	}
	return fmt.Sprintf("%s/?v=%s&encoding=%s", data.Url, apiVersion, apiEncoding), nil
}
