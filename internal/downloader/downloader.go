package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"site-mirror/internal/robots"
	"time"
)

var (
	ErrTooManyAttempts          = errors.New("too many requests")
	ErrDisallowed               = errors.New("disallowed")
	ErrCouldNotCreateDownloader = errors.New("could not create downloader")
)

const maxAttempts = 3

type Downloader struct {
	Client    *http.Client
	Robots    *robots.Robots
	UserAgent string
}

func NewDownloader(u *url.URL, userAgent string) (*Downloader, error) {
	r, err := robots.FetchRobots(u.Host)
	if err != nil {
		return nil, ErrCouldNotCreateDownloader
	}

	return &Downloader{
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
		Robots:    r,
		UserAgent: userAgent,
	}, nil
}

func (d *Downloader) Download(u *url.URL, useRobots bool) ([]byte, string, error) {
	var resp *http.Response
	var err error

	if useRobots && !d.Robots.IsAllowed(d.UserAgent, u) {
		return nil, "", ErrDisallowed
	}

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

	if resp != nil && resp.StatusCode != http.StatusOK {
		fmt.Printf("Can't download %s after attempt: %d, last status: %d\n", u.String(), attempts, resp.StatusCode)
		return nil, "", ErrTooManyAttempts
	}

	if err != nil {
		return nil, "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return respBody, resp.Header.Get("Content-Type"), nil
}
