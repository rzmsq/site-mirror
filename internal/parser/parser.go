package parser

import (
	"bytes"
	"flag"
	"net/url"
	"site-mirror/internal/config"

	"golang.org/x/net/html"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseHTML(content []byte, base *url.URL) (pages []*url.URL, resources []*url.URL, err error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, nil, err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						link, errParse := url.Parse(attr.Val)
						if errParse == nil {
							absLink := base.ResolveReference(link)
							if absLink.Host == base.Host &&
								absLink.Scheme != "mailto" &&
								absLink.Fragment == "" && // Игнор anchors
								absLink.Path != "" { // Не пустой путь
								pages = append(pages, absLink)
							}
						}
					}
				}
			case "img", "script", "link", "source":
				for _, attr := range n.Attr {
					key := attr.Key
					if key == "src" || (key == "href" && n.Data == "link") {
						res, errParse := url.Parse(attr.Val)
						if errParse == nil {
							absRes := base.ResolveReference(res)
							if absRes.Host == base.Host {
								if n.Data == "link" {
									rel := ""
									for _, a := range n.Attr {
										if a.Key == "rel" && a.Val == "stylesheet" {
											rel = a.Val
											break
										}
									}
									if rel != "stylesheet" {
										continue
									}
								}
								resources = append(resources, absRes)
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return pages, resources, nil
}

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
