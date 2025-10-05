package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const maxAttempts = 1

type Downloader struct {
	Client *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (d *Downloader) Download(u *url.URL) ([]byte, string, error) {
	var resp *http.Response
	var err error

	attempts := 0
	for attempts < maxAttempts {
		resp, err = d.Client.Get(u.String())
		if err != nil {
			attempts++
			time.Sleep(time.Second * time.Duration(attempts))
			continue
		}
		if resp.StatusCode == http.StatusOK {
			break
		}

		err = resp.Body.Close()
		if err != nil {
			return nil, "", err
		}
		attempts++
		if attempts < maxAttempts {
			time.Sleep(time.Second * time.Duration(attempts))
		}
	}

	if err != nil || resp != nil && resp.StatusCode != http.StatusOK {
		if resp != nil {
			return nil, "", fmt.Errorf("failed after %d attempts, last status: %d", attempts, resp.StatusCode)
		}
		return nil, "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return respBody, resp.Header.Get("Content-Type"), nil
}
