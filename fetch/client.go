package fetch

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func doJSON(url string, target any) error {
	resp, err := httpClient.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func newDockerClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", "/var/run/docker.sock")
			},
		},
	}
}

func doDockerJSON(client *http.Client, url string, target any) error {
	resp, err := client.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
