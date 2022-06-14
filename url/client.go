// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package url

import (
	"fmt"
	"time"

	"github.com/cavaliergopher/grab/v3"
)

var _ Client = &client{}

type Client interface {
	Download(url string, path string) error
}

func NewClient() Client {
	return &client{
		grab.NewClient(),
	}
}

type client struct {
	client *grab.Client
}

func (h client) Download(url string, path string) error {
	req, err := grab.NewRequest(path, url)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %v...\n", req.URL())
	resp := h.client.Do(req)
	fmt.Printf("HTTP response %v\n", resp.HTTPResponse.Status)

	// Start progress loop
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size(),
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		return fmt.Errorf("Download failed: %s", err)
	}

	return nil
}
