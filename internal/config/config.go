package config

import "net/url"

type Config struct {
	StartURL    *url.URL
	OutputDir   string
	Depth       int
	Concurrency int
	UseRobots   bool
}
