package parser

import (
	"flag"
	"os"
	"site-mirror/internal/config"
	"testing"
)

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
