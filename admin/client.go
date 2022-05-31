package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

var _ Client = &HttpClient{}

type Client interface {
	LoadVMs() error
	WhitelistSubnet(subnetID string) error
}

type HttpClientConfig struct {
	Endpoint string
}

type HttpClient struct {
	endpoint string
	client   http.Client
}

func NewHttpClient(config HttpClientConfig) *HttpClient {
	return &HttpClient{
		endpoint: config.Endpoint,
		client:   http.Client{},
	}
}

func (c *HttpClient) LoadVMs() error {
	body := []byte(
		fmt.Sprintf(
			`{
				"jsonrpc":"2.0",
				"id"     :1,
				"method" :"admin.loadVMs",
			}`,
		),
	)

	return c.executeHttpRequest(body)
}

// TODO handle dead node
func (c *HttpClient) WhitelistSubnet(subnetID string) error {
	body := []byte(
		fmt.Sprintf(
			`{
				"jsonrpc":"2.0",
				"id"     :1,
				"method" :"admin.whitelistSubnet",
				"params": {
					"subnetID":"%s"
				}
			}`,
			subnetID,
		),
	)

	return c.executeHttpRequest(body)
}

func (c *HttpClient) executeHttpRequest(body []byte) error {
	err := c.sendPayload(body)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		fmt.Printf("Node appears to be offline. Changes will take effect upon node restart.\n")
		// Node is offline case. This is fine, since the node will
		// automatically register the new vms upon bootstrap.
		return nil
	}

	return nil
}

func (c *HttpClient) sendPayload(body []byte) error {
	request, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Add("content-type", "application/json")

	res, err := c.client.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code %v", res.StatusCode)
	}

	return nil
}
