package parser

import (
	"flag"
	"net/url"
	"site-mirror/internal/config"
)

func ParseArgs() (*config.Config, error) {
	cfg := &config.Config{}
	var err error
	var urlRaw string

	flag.StringVar(&urlRaw, "url", "", "Start Url")
	flag.IntVar(&cfg.Depth, "depth", 5, "Depth")
	flag.StringVar(&cfg.OutputDir, "out", "./", "Output Directory")
	flag.IntVar(&cfg.Concurrency, "concurrency", 5, "Max concurrency download")
	flag.BoolVar(&cfg.UseRobots, "robots", false, "Use Robot API")
	flag.Parse()

	cfg.StartURL, err = url.Parse(urlRaw)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
