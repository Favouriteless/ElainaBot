package common

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHttpAttempts = 3

var httpClient = http.Client{Timeout: time.Second * 5}

// SendHttp signs the provided HTTP request with the client's auth headers and attempts to send it up to 3 times until a
// response or error is received. Only the final error will be returned if a response is not obtained.
func SendHttp(method string, url string, body io.Reader, headers ...string) (*http.Response, error) {
	if len(headers)%2 != 0 {
		return nil, errors.New("mismatch in header key-value count")
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}

	req.Header.Set("User-Agent", "DiscordBot (https://github.com/Favouriteless/ElainaBot, 2.0.0)")
	req.Header.Set("Authorization", "Bot "+CommonSecrets.BotToken)

	for range defaultHttpAttempts {
		var resp *http.Response
		resp, err = httpClient.Do(req)
		if resp != nil {
			return resp, nil
		}
	}
	return nil, err
}

// Get sends an HTTP GET request to the given URL signed with the bot's authorization token
func Get(url string) (*http.Response, error) {
	return SendHttp(http.MethodGet, url, nil)
}

// Post sends an HTTP POST request to the given URL signed with the bot's authorization token
func Post(url string, body io.Reader) (*http.Response, error) {
	return SendHttp(http.MethodPost, url, body, "Content-Type", "application/json")
}

// Delete sends an HTTP POST request to the given URL signed with the bot's authorization token
func Delete(url string) (*http.Response, error) {
	return SendHttp(http.MethodDelete, url, nil)
}

// Patch sends an HTTP PATCH request to the given URL signed with the bot's authorization token
func Patch(url string, body []byte) (*http.Response, error) {
	return SendHttp(http.MethodPatch, url, bytes.NewBuffer(body), "Content-Type", "application/json")
}

// Put sends an HTTP P{UT request to the given URL signed with the bot's authorization token
func Put(url string, body []byte) (*http.Response, error) {
	return SendHttp(http.MethodPut, url, bytes.NewBuffer(body), "Content-Type", "application/json")
}

func ApiUrl(parts ...string) string {
	for i, v := range parts {
		parts[i] = url.PathEscape(v)
	}
	out, err := url.JoinPath(BaseApiUrl, parts...)
	AssertIsNil(err)
	return out
}

// QueryParams takes an even-sized array of key-value pairs k,v,k,v, escapes the values and returns them as the query segment of a url
func QueryParams(parts ...string) string {
	if len(parts)%2 != 0 {
		panic(errors.New("mismatch in query param key-value pairs: " + strings.Join(parts, ",")))
	}
	out := "?"
	for i := 0; i < len(parts); i += 2 {
		out += parts[i] + "=" + url.QueryEscape(parts[i+1]) + "&"
	}
	return out[:len(out)-1] // Last character is a trailing &
}

func queryParam(key string, value string) string {
	return key + "=" + url.QueryEscape(value)
}
