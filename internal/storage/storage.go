package storage

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	BaseDir string
}

func NewStorage(baseDir string) *Storage {
	return &Storage{BaseDir: baseDir}
}

func (s *Storage) Save(u *url.URL, content []byte, contentType string) error {
	localPath := filepath.Join(s.BaseDir, u.Host)

	path := u.Path
	if path == "" || path == "/" {
		path = "index.html"
	}

	if u.RawQuery != "" {
		path = strings.Replace(path, ".", "_", -1) + "_" +
			strings.Replace(u.RawQuery, "&", "_", -1) + getExtensionFromMIME(contentType)
	} else {
		if filepath.Ext(path) == "" && strings.HasPrefix(contentType, "text/html") {
			path += ".html"
		}
	}

	localPath = filepath.Join(localPath, path)

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(localPath, content, 0644)
}

func getExtensionFromMIME(contentType string) string {
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(ct)

	switch ct {
	case "text/html":
		return ".html"
	case "text/css":
		return ".css"
	case "application/javascript", "text/javascript":
		return ".js"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ".html"
	}
}
