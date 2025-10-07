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

// GatewayPayload represents a gateway payload for discord API. Pointers are for optional fields to encode as NULL
// See: https://discord.com/developers/docs/events/gateway-events#payload-structure
type GatewayPayload struct {
	Opcode      int              `json:"op"` // Opcode of the gateway event. See: https://discord.com/developers/docs/topics/opcodes-and-status-codes#gateway-gateway-opcodes
	Data        *json.RawMessage `json:"d"`  // Data is a JSON encoded payload for the event
	SequenceNum *int             `json:"s"`  // SequenceNum of the event, used for resuming sessions and heartbeats. Will be omitted if Opcode is not 0.
	EventName   *string          `json:"t"`  // EventName of the payload held in Data. Will be omitted if Opcode is not 0.
}

func (client *Client) StartGatewaySession() error {
	if client.gateway.gatewayUrl == "" {
		if err := client.fetchGatewayUrl(3); err != nil {
			return err
		}
	}
	if err := client.connectGateway(client.gateway.gatewayUrl); err != nil {
		return err
	}
	client.identify() // After identifying, discord will either close the connection or start sending events
	return nil
}

// connectGateway attempts to initiate a websocket connection with Discord's Gateway API (https://discord.com/developers/docs/events/overview)
// and start listening for events. It will either block until the connection has closed, or return an error when something
// goes wrong during initialisation. Process is as follows:
//
// 1.) Send HTTP upgrade request
// 2.) Wait for Discord's hello payload
// 3.) Start heartbeats
func (client *Client) connectGateway(url string, closeCode chan<- int) error {
	if url == "" {
		return errors.New("cannot connect to an empty url")
	}

	log.Println("Attempting gateway connection...")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	defer conn.Close()
	conn.SetCloseHandler(func(code int, text string) error {
		closeCode <- code
		return nil
	})

	var hello Hello
	if err = readKnownGatewayEvent(conn, &hello); err != nil { // First payload sent by discord should be a Hello https://discord.com/developers/docs/events/gateway#connection-lifecycle
		return err
	}

	stopBeating := make(chan bool)
	go client.startBeating(time.Millisecond*time.Duration(hello.HeartbeatInterval), stopBeating)
	log.Println("Gateway connection established")

	client.gateway.sendBuffer = make(chan []byte, 3) // Arbitrary capacity to prevent blocking.
	stopWriter := make(chan bool)
	readerStopped := make(chan error)

	go handleGatewayWrite(conn, client.gateway.sendBuffer, stopWriter)
	go handleGatewayRead(conn, client, readerStopped)

out: // The writer will gracefully handle errors, but we can assume the connection closed if the reader stops.
	for {
		select {
		case err = <-readerStopped:
			log.Printf("Stopping gateway reader: %s", err)
			break out
		}
	}

	stopBeating <- true // Stop all subtasks when the reader is done
	stopWriter <- true
	client.gateway.sendBuffer = nil
	return nil
}

func (client *Client) startBeating(interval time.Duration, stop <-chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
			client.heartbeat()
			time.Sleep(interval)
		}
	}
}

// heartbeat creates and sends a Heartbeat payload to the Gateway API
func (client *Client) heartbeat() {
	payload := GatewayPayload{Opcode: opHeartbeat, Data: nil}
	if client.gateway.sequence != nil {
		if d, err := json.Marshal(*client.gateway.sequence); err == nil {
			payload.Data = (*json.RawMessage)(&d)
		} else {
			panic(err)
		}
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	client.gateway.sendBuffer <- encoded
	client.gateway.heartbeatAcknowledged = false
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
	client.gateway.sendBuffer <- encoded
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
				return
			}
		}
	}
}

// handleGatewayRead reads all payloads from a websocket connection and delegates their handling to relevant functions.
// No stop flag is needed as the connection will return an error when it closes anyway.
func handleGatewayRead(conn *websocket.Conn, client *Client, done chan<- error) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			done <- err
			return
		}
		payload, err := decodeGatewayPayload(msg)
		if err != nil {
			done <- err
			return
		}

		switch payload.Opcode {
		case opHeartbeat:
			log.Println("Discord requested a heartbeat")
			client.heartbeat()
		case opHeartbeatAck:
			log.Println("Heartbeat acknowledged!")
		case opDispatch:
			client.gateway.sequence = payload.SequenceNum
			log.Printf("Received event: %s", *payload.EventName)
			if payload.Data != nil {
				log.Printf(string(*payload.Data))
			}
		}
	}
}

func (client *Client) fetchGatewayUrl(attempts int) error {
	client.gateway.gatewayUrl = "" // Clear just in case

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
	client.gateway.gatewayUrl = data.Url + "/?v=" + apiVersion + "&encoding=" + apiEncoding
	if client.gateway.gatewayUrl == "" {
		return errors.New("could not fetch gateway url. this is probably an issue on discord's end")
	}

	return nil
}

// decodeGatewayPayload takes a JSON encoded []byte and decodes it into a GatewayPayload
func decodeGatewayPayload(encoded []byte) (*GatewayPayload, error) {
	var payload GatewayPayload
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// readKnownGatewayEvent waits for a websocket message and decodes it into a known type. Should only be used when the
// payload type is guaranteed, e.g. for hello.
func readKnownGatewayEvent(conn *websocket.Conn, event any) error {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	payload, err := decodeGatewayPayload(msg)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(*payload.Data, event); err != nil {
		return err
	}
	return nil
}
