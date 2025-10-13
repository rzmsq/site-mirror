package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const maxAttempts = 3

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
		fmt.Printf("Downloading %s, attempt: %d\n", u.String(), attempts)
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
		fmt.Printf("Can't download %s after attempt: %d, last status: %d\n", u.String(), attempts, resp.StatusCode)
		return nil, "", nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return respBody, resp.Header.Get("Content-Type"), nil
}
