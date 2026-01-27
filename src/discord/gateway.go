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

// Websocket Close codes as specified by https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-close-event-codes
// Codes not needed for responses have not been included (e.g. encoding errors)
const (
	closeUnknownError         = 4000
	closeNotAuthenticated     = 4003
	closeAuthenticationFailed = 4004
	closeInvalidSequence      = 4007
	closeRateLimited          = 4008
	closeTimedOut             = 4009
)

var ErrInvalidResumeUrl = errors.New("[Gateway] invalid resume url")

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

// GatewayHandle represents an active gateway connection with a disconnect and error channel. If a value is passed to
// Disconnect, the gateway will close normally. Done signals that the connection has closed. If the gateway closed
// normally, nil will be sent. Otherwise, the error causing the close will be sent.
type GatewayHandle struct {
	Done       <-chan error
	Disconnect chan<- interface{}
}

func (h *GatewayHandle) Close() {
	h.Disconnect <- true
}

// ListenGateway initializes a gateway connection and returns a GatewayHandle for controlling it.
func ListenGateway(intents int) *GatewayHandle {
	slog.Info("[Gateway] Initializing connection...")

	done := make(chan error)
	disconnect := make(chan interface{})

	go func() {
		context := &gateway{sendQueue: make(chan []byte, defaultQueueSize), intents: intents}
		disconnecting := atomic.Bool{}

		go func() {
			<-disconnect
			disconnecting.Store(true)
			context.cardiacArrest.Store(true)
			_ = context.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			context.conn.Close()
		}()

		for {
			if reconnect, err := context.connect(); !reconnect || disconnecting.Load() {
				if disconnecting.Load() {
					done <- nil
				} else {
					done <- err
				}
				return
			}
		}

	}()

	return &GatewayHandle{Done: done, Disconnect: disconnect}
}

// connect attempts to initialize a gateway connection. If resuming is true, the connection will attempt to resuming the
// last session.
func (gateway *gateway) connect() (reconnect bool, err error) {
	url, err := gateway.getConnectUrl()
	if errors.Is(err, ErrInvalidResumeUrl) {
		gateway.resuming = false
		return gateway.connect()
	} else if err != nil {
		return false, errors.New("could not fetch a gateway url: " + err.Error())
	}

	slog.Info("[Gateway] Attempting to connect:", slog.String("api_version", apiVersion), slog.String("api_encoding", apiEncoding))
	gateway.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		if os.IsTimeout(err) {
			gateway.connectUrl = ""
			return true, err
		}
		return false, err
	}
	slog.Info("[Gateway] Websocket initialized")

	var closeCode int
	gateway.conn.SetCloseHandler(func(c int, t string) error {
		closeCode = c
		return nil
	}) // Capture the close code after the wg finishes

	gateway.cardiacArrest.Store(false)
	gateway.wg.Add(1)
	go gateway.handleWriting()
	gateway.resuming, err = gateway.readUntilClosed(gateway.resuming)

	gateway.cardiacArrest.Store(true)
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
		if err = gateway.conn.ReadJSON(&payload); err != nil {
			return false, err
		}

		switch payload.Opcode {
		case opHello:
			var hello helloPayload
			if err = json.Unmarshal(*payload.Data, &hello); err != nil {
				panic(err) // Should never be hit
			}
			gateway.wg.Add(1)
			go gateway.heartbeat(time.Millisecond * time.Duration(hello.HeartbeatInterval))

			gateway.identify(resuming)
		case opDispatch:
			gateway.sequence.Store(*payload.SequenceNum)
			go gateway.dispatchEvent(*payload.EventName, *payload.Data)
		case opHeartbeat:
			gateway.sendHeartbeat()
		case opReconnect:
			slog.Info("[Gateway] Discord requested a reconnect")
			return true, nil
		case opInvalidSession:
			slog.Warn("[Gateway] Discord invalidated the session")
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
func (gateway *gateway) sendPayload(payload *gatewayPayload) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err) // Should never be hit, this state is unrecoverable
	}
	gateway.sendQueue <- encoded
}

// Identify sends either an IDENTIFY or RESUME packet depending on the value of resuming
func (gateway *gateway) identify(resume bool) {
	if resume {
		id, err := json.Marshal(resumePayload{
			Token:       application.token,
			SessionId:   gateway.sessionId,
			SequenceNum: gateway.sequence.Load(),
		})
		if err != nil {
			panic(err) // Should never be hit
		}

		gateway.sendPayload(&gatewayPayload{Opcode: opResume, Data: (*json.RawMessage)(&id)})
		slog.Info("[Gateway] Resuming connection")
		return
	}
	enc, err := json.Marshal(identifyPayload{
		Token:      application.token,
		Properties: ConnectionProperties{Os: "windows", Browser: application.name, Device: application.name},
		Intents:    gateway.intents,
	})
	if err != nil {
		panic(err) // Should never be hit
	}

	gateway.sendPayload(&gatewayPayload{Opcode: opIdentify, Data: (*json.RawMessage)(&enc)})
	slog.Info("[Gateway] Identifying connection")
}

func (gateway *gateway) sendHeartbeat() {
	data, err := json.Marshal(gateway.sequence.Load())
	if err != nil {
		panic(err) // should never be hit
	}

	gateway.sendPayload(&gatewayPayload{Opcode: opHeartbeat, Data: (*json.RawMessage)(&data)})
	gateway.heartbeatAcknowledged = false
}

func (gateway *gateway) heartbeat(interval time.Duration) {
	defer gateway.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("[Gateway] Starting heartbeats")
	gateway.sendHeartbeat()
	for {
		select {
		case <-ticker.C:
			if !gateway.heartbeatAcknowledged {
				slog.Error("[Gateway] Heartbeat not acknowledged, terminating connection and resuming")
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
	defer gateway.wg.Done()
	for {
		select {
		case payload := <-gateway.sendQueue:
			if err := gateway.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				slog.Error("[Gateway] Failed write to connection: " + err.Error()) // This should never be hit, but just in case
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
