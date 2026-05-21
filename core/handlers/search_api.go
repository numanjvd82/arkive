package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	filessvc "arkive/core/services/files"
	"arkive/core/services/shares"
	appcontext "arkive/pkg/context"
)

type searchResult struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Meta     string `json:"meta,omitempty"`
	Status   string `json:"status,omitempty"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

func APISearch(filesService *filessvc.Service, shareService *shares.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		query := strings.TrimSpace(c.Query("q"))
		if query == "" {
			c.JSON(http.StatusOK, gin.H{"results": gin.H{}})
			return
		}

		files, err := filesService.SearchCompletedUploads(c.Request.Context(), user.ID, query, 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}

		shareItems, err := shareService.SearchSharesForUser(c.Request.Context(), user.ID, query, 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": gin.H{
				"files":    mapFileResults(files),
				"shares":   mapShareResults(shareItems),
				"settings": searchSettingsResults(query),
			},
		})
	}
}

func mapFileResults(files []models.File) []searchResult {
	results := make([]searchResult, 0, len(files))
	for _, file := range files {
		url := "/api/files/" + file.ID + "/download"
		if isSearchPreviewable(file.UploadStatus) {
			url = "/files/" + file.ID + "/view"
		}
		results = append(results, searchResult{
			ID:       file.ID,
			Kind:     "file",
			Title:    "Encrypted file",
			Meta:     "Encrypted",
			URL:      url,
			Category: "Files",
		})
	}
	return results
}

func mapShareResults(items []models.ShareWithFile) []searchResult {
	now := time.Now()
	results := make([]searchResult, 0, len(items))
	for _, item := range items {
		status := "Active"
		if item.ExpiresAt != nil && !item.ExpiresAt.After(now) {
			status = "Expired"
		}
		if item.PasswordHash != nil && status == "Active" {
			status = "Restricted"
		}
		results = append(results, searchResult{
			ID:       item.ID,
			Kind:     "share",
			Title:    item.FileName,
			Status:   status,
			Meta:     "/s/" + item.Token,
			URL:      "/shares",
			Category: "Shares",
		})
	}
	return results
}

func searchSettingsResults(query string) []searchResult {
	type item struct {
		title []string
		url   string
		meta  string
	}
	settings := []item{
		{title: []string{"Instance Overview", "Instance", "Admin", "Email", "Storage", "Usage"}, url: "/settings#settings-account", meta: "Settings"},
		{title: []string{"Storage Provider", "Storage Configuration", "Provider", "Local", "S3"}, url: "/settings#settings-provider", meta: "Settings"},
		{title: []string{"Security", "Authentication", "Session", "Hardening"}, url: "/settings#settings-security", meta: "Settings"},
	}

	lower := strings.ToLower(query)
	results := []searchResult{}
	for idx, setting := range settings {
		match := false
		for _, token := range setting.title {
			if strings.Contains(strings.ToLower(token), lower) {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		results = append(results, searchResult{
			ID:       fmt.Sprintf("settings-%d", idx),
			Kind:     "setting",
			Title:    setting.title[0],
			Meta:     setting.meta,
			URL:      setting.url,
			Category: "Settings",
		})
	}
	return results
}

func isSearchPreviewable(status string) bool {
	return strings.TrimSpace(status) == "complete"
}
