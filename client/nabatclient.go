package nabatdbclient

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"bytes"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func Connect(baseURL string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) sendRequest(endpoint string, reqBody, respBody interface{}) error {
	bodyBytes, _ := json.Marshal(reqBody)
	bodyReader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK status: %d", resp.StatusCode)
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Get(key string) (string, error) {
	type request struct {
		Key string `json:"key"`
	}

	type response struct {
		Value string `json:"value"`
	}

	var resp response
	err := c.sendRequest("/get", request{Key: key}, &resp)
	return resp.Value, err
}

func (c *Client) Set(key, value string) error {
	type request struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	return c.sendRequest("/set", request{Key: key, Value: value}, nil)
}

func (c *Client) Delete(key string) error {
	type request struct {
		Key string `json:"key"`
	}

	return c.sendRequest("/delete", request{Key: key}, nil)
}
