package discord

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
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

var gateway *gatewayConnection

// gatewayConnection stores a Client's gatewayConnection connection details
type gatewayConnection struct {
	conn      *websocket.Conn // Websocket connection for the client. Only supports writing from one thread, use sendQueue to write to it
	sendQueue chan []byte     // Buffer for writing to conn, can still accept items even when disconnected

	cardiacArrest atomic.Bool    // cardiacArrest will stop all gatewayConnection child goroutines (handleWriting, heartbeat, handleReading)
	wg            sync.WaitGroup // wg will allow disconnect to block until all the child processes are finished.

	intents int // gatewayConnection intents the bot will be using https://discord.com/developers/docs/events/gateway#gateway-intents

	url       string       // URL for connecting to Discord's gateway API
	resumeUrl string       // URL for resuming a gateway connection
	sessionId string       // ID of gateway session, only applicable if resuming
	sequence  atomic.Int32 // The last sequence number the client received from the gateway

	heartbeatAcknowledged bool // Update to false when the client sends a sendHeartbeat. If discord doesn't acknowledge before the next sendHeartbeat, reconnect.
	disconnecting         bool // Stops the close handler from hanging the client is the one requesting a disconnect
}

// InitializeGateway initializes the gatewayConnection struct and attempts to send a connection request to the API.
func InitializeGateway(intents int) error {
	slog.Info("Initializing gateway connection...")
	gateway = &gatewayConnection{
		sendQueue: make(chan []byte, defaultQueueSize),
		intents:   intents,
	}
	return gateway.connect(false)
}

// CloseGateway closes the active gateway connection and waits for all child goroutines to finish.
func CloseGateway() {
	gateway.disconnect(false)
}

// gatewayPayload represents a gatewayConnection payload for discord API. Pointers are for optional fields to encode as NULL
// See: https://discord.com/developers/docs/events/gateway-events#payload-structure
type gatewayPayload struct {
	Opcode      int              `json:"op"`
	Data        *json.RawMessage `json:"d"`
	SequenceNum *int             `json:"s,omitempty"`
	EventName   *string          `json:"t,omitempty"`
}

// SendPayload sends the given gatewayPayload on the websocket connection if one is open, or errors if there is no valid
// connection or the payload fails to encode
func (gateway *gatewayConnection) SendPayload(payload *gatewayPayload) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	gateway.sendQueue <- encoded
	return nil
}

// disconnect closes the active gateway websocket and handles clean-up of background tasks. If forReconnect is true, disconnect
// will not send a disconnect frame so discord keeps the session active.
func (gateway *gatewayConnection) disconnect(forReconnect bool) {
	gateway.disconnecting = true
	gateway.cardiacArrest.Store(true) // Make sure we stop trying to write to the socket. Reader will stop itself.
	if !forReconnect {
		gateway.sequence.Store(0)
	}
	if gateway.conn != nil {

		if forReconnect {
			gateway.conn.Close()
		} else {
			_ = gateway.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		}
	}
	gateway.wg.Wait()
	gateway.conn = nil
}

// handleClosed handles clean-up of background tasks after the websocket connection was closed by discord.
func (gateway *gatewayConnection) handleClosed(code int, text string) error {
	if gateway.disconnecting {
		return nil // If we don't do this, the wait group gets called again and waits forever
	}

	reconnect := code == closeUnknownError || code == closeNotAuthenticated || code == closeInvalidSequence || code == closeRateLimited || code == closeTimedOut
	gateway.disconnect(reconnect)
	slog.Info("Gateway connection closed: " + strconv.Itoa(code))

	if reconnect {
		if err := gateway.connect(false); err != nil {
			slog.Error("Could not reconnect to gateway: " + strconv.Itoa(code) + ": " + err.Error())
		}
	}
	return nil
}

// connect attempts to initialize a gateway connection. If resume is true, the connection will attempt to resume the
// last session.
func (gateway *gatewayConnection) connect(resume bool) error {
	if gateway.conn != nil {
		return nil // Already connected, do nothing
	}

	url, err := gateway.getConnectUrl(resume)
	if err != nil {
		return err
	}

	slog.Info("Attempting to connect to gateway:", slog.String("api_version", apiVersion), slog.String("api_encoding", apiEncoding))
	gateway.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	slog.Info("Websocket created")
	gateway.conn.SetCloseHandler(gateway.handleClosed) // Handles behaviour when discord initiates a disconnect

	gateway.cardiacArrest.Store(false)
	gateway.wg.Add(2)

	go gateway.handleWriting()
	go gateway.handleReading(resume)
	return nil
}

// Identify sends either an IDENTIFY or RESUME packet depending on the value of resume
func (gateway *gatewayConnection) identify(resume bool) {
	if resume {
		id, err := json.Marshal(ResumePayload{
			Token:       application.token,
			SessionId:   gateway.sessionId,
			SequenceNum: gateway.sequence.Load(),
		})
		if err != nil {
			panic(err) // Should never be hit
		}
		if err = gateway.SendPayload(&gatewayPayload{Opcode: opResume, Data: (*json.RawMessage)(&id)}); err != nil {
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
	if err = gateway.SendPayload(&gatewayPayload{Opcode: opIdentify, Data: (*json.RawMessage)(&enc)}); err != nil {
		panic(err) // Should never be hit
	}
	slog.Info("Identifying gateway connection")
}

func (gateway *gatewayConnection) getConnectUrl(resume bool) (string, error) {
	if resume {
		if gateway.resumeUrl == "" {
			return "", errors.New("invalid resume url: " + gateway.resumeUrl)
		}
		return gateway.resumeUrl, nil
	}
	if gateway.url != "" {
		return gateway.url, nil
	}
	url, err := fetchGatewayUrl()
	if err != nil {
		return "", err
	}
	return url, nil
}

func (gateway *gatewayConnection) sendHeartbeat() {
	data, err := json.Marshal(gateway.sequence.Load())
	if err != nil {
		panic(err) // should never be hit
	}

	if err = gateway.SendPayload(&gatewayPayload{Opcode: opHeartbeat, Data: (*json.RawMessage)(&data)}); err != nil {
		slog.Error("Failed to send heartbeat: " + err.Error())
		return
	}
	gateway.heartbeatAcknowledged = false
}

func (gateway *gatewayConnection) heartbeat(interval time.Duration) {
	defer gateway.wg.Done()

	gateway.sendHeartbeat()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !gateway.heartbeatAcknowledged { // If not acknowledged, assume the connection is dead & reconnect
				defer func() { // Defer so the wait group can finish first
					gateway.disconnect(true)
					if err := gateway.connect(true); err != nil {
						slog.Error("Failed to reconnect after heartbeat: " + err.Error())
					}
				}()
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

func (gateway *gatewayConnection) handleWriting() {
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

// handleReading reads all payloads from a websocket connection and dispatches them to the event handler.
// No stop flag is needed as the connection will return an error when it closes anyway.
func (gateway *gatewayConnection) handleReading(resume bool) {
	defer gateway.wg.Done()
	for {
		if gateway.cardiacArrest.Load() {
			return
		}
		_, msg, err := gateway.conn.ReadMessage()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || gateway.disconnecting { // Little ugly but it works
				return
			}
			slog.Error("Failed to read from gateway connection: " + err.Error())
		}

		go func() {
			var payload gatewayPayload
			if err = json.Unmarshal(msg, &payload); err != nil {
				panic(err) // Should never be hit
			}

			switch payload.Opcode {
			case opHello:
				var hello HelloPayload
				if err = json.Unmarshal(*payload.Data, &hello); err != nil {
					panic(err) // Should never be hit
				}
				slog.Info("Starting heartbeats")
				gateway.wg.Add(1)
				go gateway.heartbeat(time.Millisecond * time.Duration(hello.HeartbeatInterval))
				gateway.identify(resume)
			case opDispatch:
				gateway.sequence.Store(int32(*payload.SequenceNum))
				go dispatchEvent(*payload.EventName, *payload.Data)
			case opHeartbeat:
				gateway.sendHeartbeat()
			case opReconnect:
				slog.Info("Discord requested a gateway reconnect")
				defer func() { // Defer so the wait group can finish first
					gateway.disconnect(true)
					if err := gateway.connect(true); err != nil {
						slog.Info("Failed to reconnect to gateway: " + err.Error())
					}
				}()
				return
			case opInvalidSession:
				slog.Warn("Discord invalidated the gateway session. Attempting to reconnect...")
				var resume bool
				err = json.Unmarshal(*payload.Data, &resume)
				if err != nil {
					panic(err) // Should never be hit
				}
				defer func() { // Defer so the wait group can finish first
					gateway.disconnect(resume)
					if err := gateway.connect(resume); err != nil {
						slog.Error("Failed to reconnect to gateway: " + err.Error())
					}
				}()
				return
			case opHeartbeatAck:
				gateway.heartbeatAcknowledged = true
			}
		}()
	}
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
