package discord

import (
	"encoding/json"
	"fmt"
	"io"
)

const getGatewayUrl = apiUrl + "/gateway/bot"

func (client *Client) connect() error {
	url, err := fetchGatewayUrl(client)
	if err != nil {
		return err
	}

	fmt.Println(url)
	return nil
}

func fetchGatewayUrl(client *Client) (url string, err error) {
	resp, err := client.Get(getGatewayUrl)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct{ Url string }
	if err = json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	return data.Url, nil
}
