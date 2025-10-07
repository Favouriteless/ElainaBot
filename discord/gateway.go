package discord

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const getGatewayUrl = apiUrl + "/gateway/bot"

// Gateway stores a Client's gateway connection details
type Gateway struct {
	conn       *websocket.Conn // Websocket connection for the client. This may be nil at times, always use the sendBuffer where applicable.
	sendBuffer chan []byte     // Thread-safe buffer for writing packets to a websocket connection, it also prevents losing payloads when reconnecting.

	cancelWriter  chan bool // Writing any value to this channel will stop the websocket writer routine. You shouldn't access this directly.
	cardiacArrest chan bool // Writing to cardiac arrest will make the client stop sending heartbeats

	url       string // URL for connecting to Discord's Gateway API
	resumeUrl string // URL for resuming a gateway connection
	sessionId string // ID of gateway session, only applicable if resuming
	sequence  *int   // The last sequence number the client received from gateway

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

// SendEvent sends the given GatewayEvent on the websocket connection if one is open, or errors if there is no valid
// connection or the event/payload fails to encode
func (gateway *Gateway) SendEvent(event *GatewayEvent) error {
	t := (*event).Type()
	enc, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return gateway.SendPayload(&GatewayPayload{
		Opcode:      0,
		Data:        (*json.RawMessage)(&enc),
		SequenceNum: gateway.sequence,
		EventName:   &t,
	})
}

func (gateway *Gateway) Close() {
	if gateway.conn == nil {
		return
	}
	gateway.cardiacArrest <- true
	gateway.cancelWriter <- true // Make sure we stop trying to write to the socket. Reader will close itself when it errors.
	gateway.conn.Close()
}

func (client *Client) StartGatewaySession() error {
	if client.Gateway.url == "" {
		if err := client.fetchGatewayUrl(3); err != nil {
			return err
		}
	}
	if err := client.connectGateway(client.Gateway.url); err != nil {
		return err
	}
	client.identify() // After identifying, discord will either close the connection or start sending events
	return nil
}

// connectGateway attempts to initiate a websocket connection with Discord's Gateway API (https://discord.com/developers/docs/events/overview)
// and start listening for events. It will either block until the connection has closed, or return an error when something
// goes wrong during initialisation. NOT responsible for sending an IDENTIFY or RESUME payload.
//
// 1.) Send HTTP upgrade request
// 2.) Wait for Discord's hello payload
// 3.) Start heartbeats
func (client *Client) connectGateway(url string) error {
	if url == "" {
		return errors.New("cannot connect to an empty url")
	}

	log.Println("Attempting gateway connection...")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	// TODO: Rewrite close handler

	var hello Hello
	if err = readKnownGatewayEvent(conn, &hello); err != nil { // First payload sent by discord should be a Hello https://discord.com/developers/docs/events/gateway#connection-lifecycle
		return err
	}

	client.Gateway.cardiacArrest = make(chan bool)
	client.Gateway.sendBuffer = make(chan []byte, 3) // Arbitrary capacity to prevent blocking.

	go client.Gateway.startBeating(time.Millisecond * time.Duration(hello.HeartbeatInterval))
	go handleGatewayWrite(conn, client.Gateway.sendBuffer, client.Gateway.cardiacArrest)
	go handleGatewayRead(conn, client)

	log.Println("Gateway connection established")
	return nil
}

func (gateway *Gateway) startBeating(interval time.Duration) {
	if gateway.conn == nil || gateway.cardiacArrest == nil {
		return
	}
	for {
		select {
		case <-gateway.cardiacArrest:
			return
		default:
			gateway.heartbeat()
			time.Sleep(interval)
		}
	}
}

// heartbeat creates and sends a Heartbeat payload to the Gateway API
func (gateway *Gateway) heartbeat() {
	payload := GatewayPayload{Opcode: opHeartbeat, Data: nil}
	if gateway.sequence != nil {
		if d, err := json.Marshal(*(gateway.sequence)); err == nil {
			payload.Data = (*json.RawMessage)(&d)
		} else {
			panic(err) // Should never be hit
		}
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err) // Should never be hit
	}
	gateway.sendBuffer <- encoded
	gateway.heartbeatAcknowledged = false
	log.Println("Sending heartbeat...")
}

// identify creates and sends an Identify payload to the Gateway API
func (client *Client) identify() {
	id, err := json.Marshal(Identify{
		Token: client.Token,
		Properties: ConnectionProperties{
			Os:      "windows",
			Browser: client.Name,
			Device:  client.Name,
		},
		Intents: 0,
	})
	if err != nil {
		panic(err)
	}

	encoded, err := json.Marshal(GatewayPayload{Opcode: opIdentify, Data: (*json.RawMessage)(&id)})
	if err != nil {
		panic(err)
	}
	client.Gateway.sendBuffer <- encoded
	log.Println("Sending identify...")
}

// handleGatewayWrite reads any payloads in a channel and writes them to a websocket connection. This ensures that only
// one goroutine is ever writing to the connection.
func handleGatewayWrite(conn *websocket.Conn, payloads <-chan []byte, stop <-chan bool) {
	for {
		select {
		case <-stop:
			return
		case payload := <-payloads:
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Printf("Failed to write to gateway websocket: %s", err)
				return // This should never be hit, but just in case
			}
		}
	}
}

// handleGatewayRead reads all payloads from a websocket connection and delegates their handling to relevant functions.
// No stop flag is needed as the connection will return an error when it closes anyway.
func handleGatewayRead(conn *websocket.Conn, client *Client) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var payload GatewayPayload
		if err = json.Unmarshal(msg, &payload); err != nil {
			log.Printf("Failed to decode gateway message: %s", err)
			return
		}

		switch payload.Opcode {
		case opHeartbeat:
			log.Println("Discord requested a heartbeat")
			client.Gateway.heartbeat()
		case opHeartbeatAck:
			log.Println("Heartbeat acknowledged!")
		case opDispatch:
			client.Gateway.sequence = payload.SequenceNum
			log.Printf("Received event: %s", *payload.EventName)
			if payload.Data != nil {
				log.Printf(string(*payload.Data))
			}
		}
	}
}

func (client *Client) fetchGatewayUrl(attempts int) error {
	client.Gateway.url = "" // Just in case

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
	client.Gateway.url = data.Url + "/?v=" + apiVersion + "&encoding=" + apiEncoding
	if client.Gateway.url == "" {
		return errors.New("could not fetch gateway url. this is probably an issue on discord's end")
	}

	return nil
}

// readKnownGatewayEvent waits for a websocket message and decodes it into a known type. Should only be used when the
// payload type is guaranteed, e.g. for hello.
func readKnownGatewayEvent(conn *websocket.Conn, event any) error {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	var payload GatewayPayload
	if err = json.Unmarshal(msg, &payload); err != nil {
		return err
	}
	if err = json.Unmarshal(*payload.Data, event); err != nil {
		return err
	}
	return nil
}
