package dify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	host             string
	defaultAPISecret string
	httpClient       *http.Client
}

func NewClientWithConfig(c *ClientConfig) *Client {
	var httpClient = &http.Client{}

	if c.Timeout != 0 {
		httpClient.Timeout = c.Timeout
	}
	if c.Transport != nil {
		httpClient.Transport = c.Transport
	}

	secret := c.DefaultAPISecret
	if secret == "" {
		secret = c.ApiSecretKey
	}
	return &Client{
		host:             c.Host,
		defaultAPISecret: secret,
		httpClient:       httpClient,
	}
}

func NewClient(host, defaultAPISecret string) *Client {
	return NewClientWithConfig(&ClientConfig{
		Host:             host,
		DefaultAPISecret: defaultAPISecret,
	})
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) sendJSONRequest(req *http.Request, res interface{}) error {
	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Status  int    `json:"status"`
		}
		// Tenta decodificar o corpo do erro, mas se não conseguir, usa o status http
		bodyBytes, _ := io.ReadAll(resp.Body)
		if jsonErr := json.Unmarshal(bodyBytes, &errBody); jsonErr != nil {
			return fmt.Errorf("HTTP response error: status %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return fmt.Errorf("HTTP response error: [%v] %v", errBody.Code, errBody.Message)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if res == nil {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) getHost() string {
	var host = strings.TrimSuffix(c.host, "/")
	return host
}

func (c *Client) getAPISecret() string {
	return c.defaultAPISecret
}

// Api deprecated, use API() instead
func (c *Client) Api() *API {
	return c.API()
}

func (c *Client) API() *API {
	return &API{
		c: c,
	}
}
