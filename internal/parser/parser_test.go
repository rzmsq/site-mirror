package parser

import (
	"flag"
	"net/url"
	"os"
	"site-mirror/internal/config"
	"testing"
)

func TestParser_ParseHTML(t *testing.T) {
	tests := []struct {
		name          string
		htmlContent   string
		baseURL       string
		wantPages     []string
		wantResources []string
		wantErr       bool
	}{
		{
			name: "parse links and resources",
			htmlContent: `<!DOCTYPE html>
<html>
<head>
	<link rel="stylesheet" href="/styles/main.css">
	<script src="/js/script.js"></script>
</head>
<body>
	<a href="/page1.html">Page 1</a>
	<a href="/page2.html">Page 2</a>
	<img src="/images/logo.png">
</body>
</html>`,
			baseURL:       "https://example.com",
			wantPages:     []string{"https://example.com/page1.html", "https://example.com/page2.html"},
			wantResources: []string{"https://example.com/styles/main.css", "https://example.com/js/script.js", "https://example.com/images/logo.png"},
			wantErr:       false,
		},
		{
			name: "ignore external links",
			htmlContent: `<!DOCTYPE html>
<html>
<body>
	<a href="https://external.com/page">External</a>
	<a href="/internal.html">Internal</a>
	<img src="https://cdn.com/image.png">
	<img src="/local.png">
</body>
</html>`,
			baseURL:       "https://example.com",
			wantPages:     []string{"https://example.com/internal.html"},
			wantResources: []string{"https://example.com/local.png"},
			wantErr:       false,
		},
		{
			name: "ignore anchors and mailto",
			htmlContent: `<!DOCTYPE html>
<html>
<body>
	<a href="#section1">Section 1</a>
	<a href="mailto:test@example.com">Email</a>
	<a href="/page.html">Valid Page</a>
</body>
</html>`,
			baseURL:       "https://example.com",
			wantPages:     []string{"https://example.com/page.html"},
			wantResources: []string{},
			wantErr:       false,
		},
		{
			name: "ignore non-stylesheet links",
			htmlContent: `<!DOCTYPE html>
<html>
<head>
	<link rel="icon" href="/favicon.ico">
	<link rel="stylesheet" href="/style.css">
</head>
</html>`,
			baseURL:       "https://example.com",
			wantPages:     []string{},
			wantResources: []string{"https://example.com/style.css"},
			wantErr:       false,
		},
		{
			name: "empty HTML",
			htmlContent: `<!DOCTYPE html>
<html>
<body></body>
</html>`,
			baseURL:       "https://example.com",
			wantPages:     []string{},
			wantResources: []string{},
			wantErr:       false,
		},
		{
			name:          "invalid HTML",
			htmlContent:   `<html><body><a href="/page">`,
			baseURL:       "https://example.com",
			wantPages:     []string{"https://example.com/page"},
			wantResources: []string{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			baseURL, _ := url.Parse(tt.baseURL)

			pages, resources, err := p.ParseHTML([]byte(tt.htmlContent), baseURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check pages
			if len(pages) != len(tt.wantPages) {
				t.Errorf("ParseHTML() got %d pages, want %d", len(pages), len(tt.wantPages))
			}
			for i, wantPage := range tt.wantPages {
				if i >= len(pages) {
					break
				}
				if pages[i].String() != wantPage {
					t.Errorf("ParseHTML() page[%d] = %v, want %v", i, pages[i].String(), wantPage)
				}
			}

			// Check resources
			if len(resources) != len(tt.wantResources) {
				t.Errorf("ParseHTML() got %d resources, want %d", len(resources), len(tt.wantResources))
			}
			for i, wantRes := range tt.wantResources {
				if i >= len(resources) {
					break
				}
				if resources[i].String() != wantRes {
					t.Errorf("ParseHTML() resource[%d] = %v, want %v", i, resources[i].String(), wantRes)
				}
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		checks  func(t *testing.T, cfg *config.Config)
	}{
		{
			name:    "valid arguments with all flags",
			args:    []string{"-url", "https://example.com", "-depth", "3", "-out", "/tmp", "-concurrency", "10", "-robots"},
			wantErr: false,
			checks: func(t *testing.T, cfg *config.Config) {
				if cfg.StartURL.String() != "https://example.com" {
					t.Errorf("expected URL https://example.com, got %s", cfg.StartURL.String())
				}
				if cfg.Depth != 3 {
					t.Errorf("expected depth 3, got %d", cfg.Depth)
				}
				if cfg.OutputDir != "/tmp" {
					t.Errorf("expected output dir /tmp, got %s", cfg.OutputDir)
				}
				if cfg.Concurrency != 10 {
					t.Errorf("expected concurrency 10, got %d", cfg.Concurrency)
				}
				if !cfg.UseRobots {
					t.Error("expected UseRobots to be true")
				}
			},
		},
		{
			name:    "default values",
			args:    []string{"-url", "https://example.com"},
			wantErr: false,
			checks: func(t *testing.T, cfg *config.Config) {
				if cfg.Depth != 5 {
					t.Errorf("expected default depth 5, got %d", cfg.Depth)
				}
				if cfg.OutputDir != "./" {
					t.Errorf("expected default output dir ./, got %s", cfg.OutputDir)
				}
				if cfg.Concurrency != 5 {
					t.Errorf("expected default concurrency 5, got %d", cfg.Concurrency)
				}
				if cfg.UseRobots {
					t.Error("expected UseRobots to be false by default")
				}
			},
		},
		{
			name:    "invalid URL",
			args:    []string{"-url", "://invalid-url"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			os.Args = append([]string{"cmd"}, tt.args...)
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := ParseArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checks != nil {
				tt.checks(t, cfg)
			}
		})
	}
}
