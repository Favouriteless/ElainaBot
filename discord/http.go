package discord

import (
	"bytes"
	"net/http"
)

// SendHttpAuth signs the provided HTTP request with the client's auth headers and forwards it to SendHttp
func (client *Client) SendHttpAuth(req *http.Request, attempts int) (*http.Response, error) {
	return client.SendHttp(client.setAuthHeaders(req), attempts)
}

// SendHttp attempts to send the given HTTP request up to N times until a response or error is received. Only the final
// error will be returned if a response is not obtained.
func (client *Client) SendHttp(req *http.Request, attempts int) (*http.Response, error) {
	var resp *http.Response
	var err error
	for range attempts {
		resp, err = client.Http.Do(req)
		if resp != nil {
			return resp, nil
		}
	}
	return nil, err
}

// Get sends an HTTP GET request to the given URL signed with the bot's authorisation token
func (client *Client) Get(url string, attempts int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return client.SendHttpAuth(req, attempts)
}

// Post sends an HTTP POST request to the given URL signed with the bot's authorisation token
func (client *Client) Post(url string, body []byte, attempts int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.SendHttpAuth(req, attempts)
}

// Delete sends an HTTP POST request to the given URL signed with the bot's authorisation token
func (client *Client) Delete(url string, attempts int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return client.SendHttpAuth(req, attempts)
}

// Patch sends an HTTP PATCH request to the given URL signed with the bot's authorisation token
func (client *Client) Patch(url string, body []byte, attempts int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.SendHttpAuth(req, attempts)
}

// setAuthHeaders signs the given HTTP request with the client's user agent and auth token
func (client *Client) setAuthHeaders(req *http.Request) *http.Request {
	req.Header.Set("User-Agent", "DiscordBot (https://github.com/Favouriteless/ElainaBot, 2.0.0)")
	req.Header.Set("Authorization", "Bot "+client.Token)
	return req
}
