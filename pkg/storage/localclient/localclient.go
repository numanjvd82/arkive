package localclient

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"arkive/pkg/header"
)

type RootFunc func(context.Context) (string, error)

type Client struct {
	rootFunc RootFunc
	mu       sync.Mutex
	tokens   map[string]tokenEntry
}

type tokenEntry struct {
	Key         string
	Filename    string
	Disposition string
	ContentType string
	ExpiresAt   time.Time
	Download    bool
}

func New(rootFunc RootFunc) *Client {
	return &Client{
		rootFunc: rootFunc,
		tokens:   map[string]tokenEntry{},
	}
}

func (c *Client) PresignUpload(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("key is required")
	}
	token, err := newToken()
	if err != nil {
		return "", err
	}
	c.storeToken(token, tokenEntry{
		Key:         key,
		ContentType: contentType,
		ExpiresAt:   time.Now().Add(expires),
	})
	return "/local-storage/upload/" + token, nil
}

func (c *Client) PresignDownload(ctx context.Context, key, filename, disposition string, expires time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("key is required")
	}
	token, err := newToken()
	if err != nil {
		return "", err
	}
	c.storeToken(token, tokenEntry{
		Key:         key,
		Filename:    filename,
		Disposition: disposition,
		ExpiresAt:   time.Now().Add(expires),
		Download:    true,
	})
	return "/local-storage/download/" + token, nil
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	path, err := c.objectPath(ctx, key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (c *Client) ServeUpload(w http.ResponseWriter, r *http.Request, token string) {
	entry, ok := c.consumeToken(token, false)
	if !ok {
		http.NotFound(w, r)
		return
	}
	path, err := c.objectPath(r.Context(), entry.Key)
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	if _, err := io.Copy(file, r.Body); err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Client) ServeDownload(w http.ResponseWriter, r *http.Request, token string) {
	entry, ok := c.getToken(token, true)
	if !ok {
		http.NotFound(w, r)
		return
	}
	path, err := c.objectPath(r.Context(), entry.Key)
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	if contentDisposition := header.BuildContentDisposition(entry.Filename, entry.Disposition); contentDisposition != "" {
		w.Header().Set("Content-Disposition", contentDisposition)
	}
	http.ServeFile(w, r, path)
}

func (c *Client) storeToken(token string, entry tokenEntry) {
	if entry.ExpiresAt.IsZero() {
		entry.ExpiresAt = time.Now().Add(15 * time.Minute)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[token] = entry
}

func (c *Client) consumeToken(token string, download bool) (tokenEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.tokens[token]
	if !ok || entry.Download != download || time.Now().After(entry.ExpiresAt) {
		delete(c.tokens, token)
		return tokenEntry{}, false
	}
	delete(c.tokens, token)
	return entry, true
}

func (c *Client) getToken(token string, download bool) (tokenEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.tokens[token]
	if !ok || entry.Download != download || time.Now().After(entry.ExpiresAt) {
		delete(c.tokens, token)
		return tokenEntry{}, false
	}
	return entry, true
}

func (c *Client) objectPath(ctx context.Context, key string) (string, error) {
	root, err := c.rootFunc(ctx)
	if err != nil {
		return "", err
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("local storage path is required")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	cleanKey := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(key, "/")))
	if cleanKey == "." || strings.HasPrefix(cleanKey, "..") {
		return "", errors.New("invalid object key")
	}
	path := filepath.Join(rootAbs, cleanKey)
	if path != rootAbs && !strings.HasPrefix(path, rootAbs+string(os.PathSeparator)) {
		return "", errors.New("invalid object key")
	}
	return path, nil
}

func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
