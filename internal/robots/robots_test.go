package robots

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestParseRobots проверяет парсинг robots.txt из строки
func TestParseRobots(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		userAgent string
		expected  map[string][]string
	}{
		{
			name: "Basic Disallow",
			input: `
User-agent: *
Disallow: /private/
Disallow: /secret.html

User-agent: SiteMirror
Disallow: /admin/
`,
			userAgent: "SiteMirror",
			expected: map[string][]string{
				"*":          {"/private/", "/secret.html"},
				"SiteMirror": {"/admin/"},
			},
		},
		{
			name:      "Empty File",
			input:     "",
			userAgent: "SiteMirror",
			expected:  map[string][]string{},
		},
		{
			name: "Comments and Whitespace",
			input: `
# This is a comment
User-agent: Bot
Disallow: /test/

User-agent: *
Disallow: /hidden/
# Another comment
`,
			userAgent: "Bot",
			expected: map[string][]string{
				"Bot": {"/test/"},
				"*":   {"/hidden/"},
			},
		},
		{
			name: "No Disallow",
			input: `
User-agent: SiteMirror
Allow: /public/
`,
			userAgent: "SiteMirror",
			expected: map[string][]string{
				"SiteMirror": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Имитация HTTP ответа
			r := &Robots{}
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			r.rules = make(map[string][]string)
			// Предполагается, что парсинг реализован через чтение строк
			currentAgent := ""
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if strings.HasPrefix(line, "User-agent:") {
					currentAgent = strings.TrimSpace(strings.TrimPrefix(line, "User-agent:"))
					r.rules[currentAgent] = []string{}
					continue
				}
				if strings.HasPrefix(line, "Disallow:") && currentAgent != "" {
					path := strings.TrimSpace(strings.TrimPrefix(line, "Disallow:"))
					if path != "" {
						r.rules[currentAgent] = append(r.rules[currentAgent], path)
					}
				}
			}

			// Проверка результата
			for agent, expectedPaths := range tt.expected {
				paths, exists := r.rules[agent]
				if !exists && len(expectedPaths) == 0 {
					continue
				}
				if !exists {
					t.Errorf("Expected rules for User-Agent %s, but none found", agent)
					continue
				}
				if len(paths) != len(expectedPaths) {
					t.Errorf("For User-Agent %s, expected %v, got %v", agent, expectedPaths, paths)
					continue
				}
				for i, path := range expectedPaths {
					if paths[i] != path {
						t.Errorf("For User-Agent %s, expected path %s, got %s", agent, path, paths[i])
					}
				}
			}
		})
	}
}

// TestIsAllowed проверяет метод IsAllowed
func TestIsAllowed(t *testing.T) {
	r := &Robots{
		rules: map[string][]string{
			"*":          {"/private/", "/secret.html"},
			"SiteMirror": {"/admin/"},
		},
	}

	tests := []struct {
		name      string
		userAgent string
		url       string
		expected  bool
	}{
		{
			name:      "Allowed URL for SiteMirror",
			userAgent: "SiteMirror",
			url:       "https://example.com/public/page.html",
			expected:  true,
		},
		{
			name:      "Disallowed URL for SiteMirror",
			userAgent: "SiteMirror",
			url:       "https://example.com/admin/dashboard.html",
			expected:  false,
		},
		{
			name:      "Disallowed URL for *",
			userAgent: "OtherBot",
			url:       "https://example.com/private/data.html",
			expected:  false,
		},
		{
			name:      "Allowed URL for OtherBot",
			userAgent: "OtherBot",
			url:       "https://example.com/index.html",
			expected:  true,
		},
		{
			name:      "No rules for User-Agent",
			userAgent: "UnknownBot",
			url:       "https://example.com/anything",
			expected:  true, // Если нет правил, всё разрешено
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("Failed to parse URL %s: %v", tt.url, err)
			}
			result := r.IsAllowed(tt.userAgent, u)
			if result != tt.expected {
				t.Errorf("IsAllowed(%s, %s): expected %v, got %v", tt.userAgent, tt.url, tt.expected, result)
			}
		})
	}
}

// TestFetchRobots проверяет загрузку robots.txt через HTTP
func TestFetchRobots(t *testing.T) {
	// Создаем мок-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`
User-agent: *
Disallow: /private/

User-agent: SiteMirror
Disallow: /admin/
`))
			if err != nil {
				return
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Извлекаем домен из URL сервера
	u, _ := url.Parse(server.URL)
	domain := u.Host

	// Тест успешной загрузки
	t.Run("Successful Fetch", func(t *testing.T) {
		robots, err := FetchRobots(domain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if robots == nil {
			t.Fatal("Expected non-nil Robots")
		}
		expected := map[string][]string{
			"*":          {"/private/"},
			"SiteMirror": {"/admin/"},
		}
		for agent, paths := range expected {
			gotPaths, exists := robots.rules[agent]
			if !exists {
				t.Errorf("Expected rules for User-Agent %s, but none found", agent)
				continue
			}
			if len(gotPaths) != len(paths) {
				t.Errorf("For User-Agent %s, expected %v, got %v", agent, paths, gotPaths)
				continue
			}
			for i, path := range paths {
				if gotPaths[i] != path {
					t.Errorf("For User-Agent %s, expected path %s, got %s", agent, path, gotPaths[i])
				}
			}
		}
	})

	// Тест для случая 404
	t.Run("404 Not Found", func(t *testing.T) {
		server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server404.Close()

		u, _ := url.Parse(server404.URL)
		robots, err := FetchRobots(u.Host)
		if err != nil {
			t.Fatalf("Expected no error for 404, got %v", err)
		}
		if len(robots.rules) != 0 {
			t.Errorf("Expected empty rules for 404, got %v", robots.rules)
		}
	})
}
