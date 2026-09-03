package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func doJSON(url string, target any) error {
	start := time.Now()
	resp, err := httpClient.Get(url) //nolint:noctx
	lat := time.Since(start)

	ex := Exchange{
		Kind:    "http",
		Method:  "GET",
		URL:     url,
		Request: "(none)",
		Latency: lat,
	}
	if err != nil {
		ex.Error = err.Error()
		recordTrace(ex)
		return err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		ex.Error = readErr.Error()
		recordTrace(ex)
		return readErr
	}
	ex.Response = truncateExchangeResponse(compactJSONFull(body))
	ex.OK = resp.StatusCode == 200
	if !ex.OK {
		ex.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		recordTrace(ex)
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	recordTrace(ex)

	if decodeErr := json.Unmarshal(body, target); decodeErr != nil {
		return decodeErr
	}
	return nil
}
