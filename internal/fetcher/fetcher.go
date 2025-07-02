package fetcher

import (
	"io"
	"net/http"
	"time"
)

func FetchHTML(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2_000_000))
	if err != nil {
		return "", err
	}

	return string(body), nil
}
