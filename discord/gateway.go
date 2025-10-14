package discord

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const getGatewayUrl = apiUrl + "/gateway/bot"

// Websocket close codes as specified by https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-close-event-codes
// Codes not needed for responses (e.g. encoding errors) have not been included.
const (
	closeUnknownError         = 4000
	closeNotAuthenticated     = 4003
	closeAuthenticationFailed = 4004
	closeInvalidSequence      = 4007
	closeRateLimited          = 4008
	closeTimedOut             = 4009
)

// Gateway stores a Client's gateway connection details
type Gateway struct {
	conn       *websocket.Conn // Websocket connection for the client. This may be nil at times, always use the sendBuffer where applicable.
	sendBuffer chan []byte     // Thread-safe buffer for writing packets to a websocket connection, it also prevents losing payloads when reconnecting.

	writersBlock  chan bool // writersBlock will stop the websocket writer routine. Don't access this directly.
	cardiacArrest chan bool // cardiacArrest will make the client stop sending heartbeats. Don't access this directly.

	intents int // Gateway intents the bot will be using https://discord.com/developers/docs/events/gateway#gateway-intents

	url        string // URL for connecting to Discord's Gateway API
	resumeUrl  string // URL for resuming a gateway connection
	sessionId  string // ID of gateway session, only applicable if resuming
	sequence   *int   // The last sequence number the client received from gateway
	sequenceMu sync.Mutex

	heartbeatAcknowledged bool // Set to false when the client sends a heartbeat. If discord doesn't acknowledge before the next heartbeat, we reconnect. No mutex needed as booleans don't tear and there's a large interval
}

// GatewayPayload represents a gateway payload for discord API. Pointers are for optional fields to encode as NULL
// See: https://discord.com/developers/docs/events/gateway-events#payload-structure
type GatewayPayload struct {
	Opcode      int              `json:"op"` // Opcode of the gateway event. See: https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-opcodes
	Data        *json.RawMessage `json:"d"`  // Data is a JSON encoded payload for the event
	SequenceNum *int             `json:"s"`  // SequenceNum of the event, used for resuming sessions and heartbeats. Will be omitted if Opcode is not 0.
	EventName   *string          `json:"t"`  // EventName of the payload held in Data. Will be omitted if Opcode is not 0.
}

// SendPayload sends the given GatewayPayload on the websocket connection if one is open, or errors if there is no valid
// connection or the payload fails to encode
func (gateway *Gateway) SendPayload(payload *GatewayPayload) error {
	if gateway.sendBuffer == nil || gateway.conn == nil {
		return errors.New("cannot send a gateway payload without a valid connection being open")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	gateway.sendBuffer <- encoded
	return nil
}

// CloseGateway closes the active gateway websocket and handles cleanup of background tasks.
func (client *Client) CloseGateway() {
	gateway := client.Gateway
	if gateway.conn != nil {
		gateway.conn.Close()
	}
	gateway.cardiacArrest <- true
	gateway.writersBlock <- true // Make sure we stop trying to write to the socket. Reader will stop itself.
	gateway.conn = nil
}

// handleClosed handles cleanup of background tasks after the websocket connection was closed by discord.
func (client *Client) handleClosed(code int) error {
	client.Gateway.cardiacArrest <- true
	client.Gateway.writersBlock <- true
	client.Gateway.conn = nil

	log.Printf("Discord closed the gateway connection: %v", code)

	switch code {
	case closeUnknownError: // Don't really like this but what else can you do
		client.ReconnectGateway(true)
	case closeNotAuthenticated:
		client.ReconnectGateway(false)
	case closeAuthenticationFailed:
		client.ReconnectGateway(false)
	case closeInvalidSequence:
		client.ReconnectGateway(false)
	case closeRateLimited:
		client.ReconnectGateway(true)
	case closeTimedOut:
		client.ReconnectGateway(false)
	}

	return nil
}

// ConnectGateway attempts to initiate a websocket connection with Discord's Gateway API (https://discord.com/developers/docs/events/overview)
// and start listening for events. It will either block until the connection has closed, or return an error when something
// goes wrong during initialisation.
//
// 1.) Send HTTP upgrade request
// 2.) Wait for Discord's hello payload
// 3.) Start background tasks (writer, reader, heartbeat)
// 4.) Send IDENTIFY or RESUME
func (client *Client) ConnectGateway() error {
	return client.connectGateway(false)
}

func (client *Client) connectGateway(resume bool) (err error) {
	gateway := client.Gateway
	if gateway.conn != nil {
		return nil // Already connected, do nothing
	}

	var url string
	if resume {
		url = client.Gateway.resumeUrl
		if url == "" {
			return errors.New("cannot resume without a valid resume url")
		}
	} else {
		if client.Gateway.url == "" {
			if err = client.fetchGatewayUrl(3); err != nil {
				return err
			}
		}
		url = client.Gateway.url
	}

	log.Println("Attempting gateway connection...")
	gateway.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	defer func() { // Ensure we cancel the write/beater routines & disconnect if handshake fails.
		if err != nil {
			client.CloseGateway()
		}
	}()
	gateway.conn.SetCloseHandler(func(code int, text string) error { return client.handleClosed(code) })

	// First payload sent by discord should be a HELLO https://discord.com/developers/docs/events/gateway#connection-lifecycle
	// UPDATE: It can also in some instances be INVALID SESSION (thanks discord, very glad you documented this (you didn't) )
	payload, err := readPayload(gateway.conn)
	if err != nil {
		return err
	}

	if payload.Opcode == opInvalidSession {
		var reResume bool
		err = json.Unmarshal(*payload.Data, &reResume)
		defer client.ReconnectGateway(reResume) // We don't need to close here as it's already deferred above
		return errors.New("discord invalidated the session")
	}

	var hello HelloPayload
	if err = json.Unmarshal(*payload.Data, &hello); err != nil {
		return err
	}
	client.heartbeat() // Send initial hello heartbeat

	// Start background tasks for websockets
	gateway.cardiacArrest = make(chan bool)
	gateway.writersBlock = make(chan bool)

	go client.startBeating(time.Millisecond * time.Duration(hello.HeartbeatInterval))
	go handleWriting(gateway.conn, gateway.sendBuffer, gateway.writersBlock)
	go handleReading(gateway.conn, client)

	// Send IDENTIFY or RESUME handshake. The result of these will be handled by handleReading
	if resume {
		gateway.sequenceMu.Lock()
		id, err := json.Marshal(ResumePayload{Token: client.Token, SessionId: gateway.sessionId, SequenceNum: *gateway.sequence})
		gateway.sequenceMu.Unlock()
		if err != nil {
			panic(err) // Should never be hit
		}
		if err = gateway.SendPayload(&GatewayPayload{Opcode: opResume, Data: (*json.RawMessage)(&id)}); err != nil {
			panic(err) // Should never be hit
		}
		log.Println("Sending resume...")
	} else {
		res, err := json.Marshal(IdentifyPayload{
			Token:      client.Token,
			Properties: ConnectionProperties{Os: "windows", Browser: client.Name, Device: client.Name},
			Intents:    gateway.intents,
		})
		if err != nil {
			panic(err) // Should never be hit
		}
		if err = gateway.SendPayload(&GatewayPayload{Opcode: opIdentify, Data: (*json.RawMessage)(&res)}); err != nil {
			panic(err) // Should never be hit
		}
		log.Println("Sending identify...")
	}

	log.Println("Gateway connection established")
	return nil
}

// ReconnectGateway will attempt to restore the client's gateway websocket connection. If resuming, one resume attempt
// will be performed, then an identify attempt after. If not, up to two identify attempts will be performed (to account
// for an outdated url)
func (client *Client) ReconnectGateway(resume bool) {
	if client.Gateway.conn != nil {
		client.CloseGateway()
	}
	if err := client.connectGateway(resume); err == nil {
		return
	}
	_ = client.connectGateway(false)
}

func (client *Client) startBeating(interval time.Duration) {
	if client.Gateway.conn == nil || client.Gateway.cardiacArrest == nil {
		return
	}
	for {
		select {
		case <-client.Gateway.cardiacArrest:
			return
		default:
			time.Sleep(interval)
			if !client.Gateway.heartbeatAcknowledged { // If not acknowledged, assume the connection is dead & reconnect
				client.ReconnectGateway(true)
			} else {
				client.heartbeat()
			}
		}
	}
}

func (client *Client) heartbeat() {
	payload := GatewayPayload{Opcode: opHeartbeat, Data: nil}
	client.Gateway.sequenceMu.Lock()
	if client.Gateway.sequence != nil {
		if d, err := json.Marshal(*(client.Gateway.sequence)); err == nil {
			payload.Data = (*json.RawMessage)(&d)
		} else {
			panic(err) // Should never be hit
		}
	}
	client.Gateway.sequenceMu.Unlock()
	if err := client.Gateway.SendPayload(&payload); err != nil {
		return
	}
	client.Gateway.heartbeatAcknowledged = false
	log.Println("Sending heartbeat...")
}

// handleWriting reads any payloads in a channel and writes them to a websocket connection. This ensures that only
// one goroutine is ever writing to the connection.
func handleWriting(conn *websocket.Conn, payloads <-chan []byte, stop <-chan bool) {
	for {
		select {
		case <-stop:
			return
		case payload := <-payloads:
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Printf("Failed to write to gateway websocket: %s", err) // This should never be hit, but just in case
			}
		}
	}
}

// handleReading reads all payloads from a websocket connection and delegates their handling to relevant functions.
// No stop flag is needed as the connection will return an error when it closes anyway.
func handleReading(conn *websocket.Conn, client *Client) {
	gateway := client.Gateway
	for {
		payload, err := readPayload(conn)
		if err != nil {
			log.Printf("Failed to read gateway message: %s", err)
			return
		}

		switch payload.Opcode {
		case opDispatch:
			log.Printf("Received event: %s", *payload.EventName)
			gateway.sequenceMu.Lock()
			gateway.sequence = payload.SequenceNum
			gateway.sequenceMu.Unlock()
			client.Events.dispatchEvent(*payload.EventName, *payload.Data)
		case opHeartbeat:
			log.Println("Discord requested a heartbeat")
			client.heartbeat()
		case opReconnect:
			log.Println("Discord requested a reconnect")
			client.ReconnectGateway(true)
		case opInvalidSession:
			log.Println("Discord invalidated the session")
			var resume bool
			_ = json.Unmarshal(*payload.Data, &resume) // We want to reconnect regardless, it'll just start a new session instead.
			client.ReconnectGateway(resume)
		case opHeartbeatAck:
			gateway.heartbeatAcknowledged = true
			log.Println("Heartbeat acknowledged")
		} // HELLO opcode can also be received but is handled by connectGateway before the reader is started
	}
}

func (client *Client) fetchGatewayUrl(attempts int) error {
	resp, err := client.Get(getGatewayUrl, attempts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data struct{ Url string }
	if err = json.Unmarshal(body, &data); err != nil {
		return err
	}
	client.Gateway.url = fmt.Sprintf("%s/?v=%s&encoding=%s", data.Url, apiVersion, apiEncoding)
	if client.Gateway.url == "" {
		return errors.New("could not fetch gateway url. this is probably an issue on discord's end")
	}
	return nil
}

// readPayload is a blocking operation which waits for a websocket to receive data and parses it as a GatewayPayload
func readPayload(conn *websocket.Conn) (*GatewayPayload, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	var payload GatewayPayload
	if err = json.Unmarshal(msg, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
