package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type Client interface {
	LoadVMs() error
	WhitelistSubnet(subnetID string) error
}

var _ Client = &HttpClient{}

type HttpClientConfig struct {
	Endpoint string
}

type HttpClient struct {
	endpoint string
}

func NewHttpClient(config HttpClientConfig) *HttpClient {
	return &HttpClient{
		endpoint: config.Endpoint,
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

	err := c.executeHttpRequest(body)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		// Node is offline case. This is fine, since the node will
		// automatically register the new vms upon bootstrap.
		return nil
	}

	return err
}

//TODO handle dead node
func (c *HttpClient) WhitelistSubnet(subnetID string) error {
	body := []byte(
		fmt.Sprintf(
			`{
				"jsonrpc":"2.0",
				"id"     :1,
				"method" :"admin.whitelistSubnet",
				"params": {
					"subnetID":"%sasdfasdf"
				}
			}`,
			subnetID,
		),
	)

	return c.executeHttpRequest(body)
}

func (c *HttpClient) executeHttpRequest(body []byte) error {
	request, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Add("content-type", "application/json")

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code %s", res.StatusCode)
	}

	return nil
}
