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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"arkive/pkg/header"
	"arkive/pkg/storage"
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
	UploadID    string
	PartNumber  int32
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

func (c *Client) CreateMultipartUpload(ctx context.Context, key, contentType string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("key is required")
	}
	return newToken()
}

func (c *Client) PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int32, expires time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("key is required")
	}
	if strings.TrimSpace(uploadID) == "" {
		return "", errors.New("uploadID is required")
	}
	if partNumber <= 0 {
		return "", errors.New("partNumber must be greater than 0")
	}
	token, err := newToken()
	if err != nil {
		return "", err
	}
	c.storeToken(token, tokenEntry{
		Key:        key,
		UploadID:   uploadID,
		PartNumber: partNumber,
		ExpiresAt:  time.Now().Add(expires),
	})
	return "/local-storage/upload/" + token, nil
}

func (c *Client) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("key is required")
	}
	if strings.TrimSpace(uploadID) == "" {
		return errors.New("uploadID is required")
	}
	if len(parts) == 0 {
		return errors.New("parts are required")
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	finalPath, err := c.objectPath(ctx, key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o700); err != nil {
		return err
	}

	output, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer output.Close()

	for _, part := range parts {
		partPath, err := c.multipartPartPath(ctx, uploadID, part.PartNumber)
		if err != nil {
			return err
		}
		input, err := os.Open(partPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(output, input); err != nil {
			input.Close()
			return err
		}
		if err := input.Close(); err != nil {
			return err
		}
	}

	return os.RemoveAll(c.multipartDir(ctx, uploadID))
}

func (c *Client) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	if strings.TrimSpace(uploadID) == "" {
		return errors.New("uploadID is required")
	}
	if err := os.RemoveAll(c.multipartDir(ctx, uploadID)); err != nil && !errors.Is(err, os.ErrNotExist) {
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
	if entry.UploadID != "" && entry.PartNumber > 0 {
		path, err = c.multipartPartPath(r.Context(), entry.UploadID, entry.PartNumber)
		if err != nil {
			http.Error(w, "storage unavailable", http.StatusInternalServerError)
			return
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			http.Error(w, "storage unavailable", http.StatusInternalServerError)
			return
		}
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

func (c *Client) multipartDir(ctx context.Context, uploadID string) string {
	root, err := c.rootFunc(ctx)
	if err != nil {
		return ""
	}
	rootAbs, err := filepath.Abs(strings.TrimSpace(root))
	if err != nil {
		return ""
	}
	return filepath.Join(rootAbs, ".multipart", uploadID)
}

func (c *Client) multipartPartPath(ctx context.Context, uploadID string, partNumber int32) (string, error) {
	if strings.TrimSpace(uploadID) == "" {
		return "", errors.New("uploadID is required")
	}
	if partNumber <= 0 {
		return "", errors.New("partNumber must be greater than 0")
	}
	dir := c.multipartDir(ctx, uploadID)
	if dir == "" {
		return "", errors.New("storage unavailable")
	}
	return filepath.Join(dir, "part-"+leftPadInt(int(partNumber), 6)+".bin"), nil
}

func leftPadInt(value, width int) string {
	raw := strconv.Itoa(value)
	if len(raw) >= width {
		return raw
	}
	return strings.Repeat("0", width-len(raw)) + raw
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
