package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"arkive/core/models"
)

const metadataTimeout = 30 * time.Second

type videoMetadata struct {
	Width           int
	Height          int
	DurationSeconds int64
}

type ffprobeOutput struct {
	Streams []struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Duration string `json:"duration"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func (s *Service) scheduleVideoMetadata(file models.File) {
	if !isVideoContentType(file.ContentType) {
		return
	}
	if file.VideoWidth > 0 || file.VideoHeight > 0 || file.VideoDurationSeconds > 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), metadataTimeout)
		defer cancel()
		_ = s.collectVideoMetadata(ctx, file)
	}()
}

func (s *Service) collectVideoMetadata(ctx context.Context, file models.File) error {
	url, err := s.storage.PresignDownload(ctx, file.ObjectKey, file.Filename, "inline", s.downloadExpire)
	if err != nil {
		return err
	}

	meta, err := probeVideoMetadata(ctx, url)
	if err != nil {
		return err
	}
	if meta.Width == 0 && meta.Height == 0 && meta.DurationSeconds == 0 {
		return nil
	}

	return s.fileRepo.UpdateVideoMetadata(ctx, s.db, file.ID, meta.Width, meta.Height, meta.DurationSeconds)
}

func probeVideoMetadata(ctx context.Context, url string) (videoMetadata, error) {
	if strings.TrimSpace(url) == "" {
		return videoMetadata{}, errors.New("missing url")
	}

	cmd := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		"-show_entries", "format=duration",
		"-of", "json",
		url,
	)
	output, err := cmd.Output()
	if err != nil {
		return videoMetadata{}, err
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return videoMetadata{}, err
	}

	meta := videoMetadata{}
	if len(parsed.Streams) > 0 {
		meta.Width = parsed.Streams[0].Width
		meta.Height = parsed.Streams[0].Height
		meta.DurationSeconds = parseDurationSeconds(parsed.Streams[0].Duration)
	}
	if meta.DurationSeconds == 0 {
		meta.DurationSeconds = parseDurationSeconds(parsed.Format.Duration)
	}

	return meta, nil
}

func parseDurationSeconds(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	if parsed <= 0 {
		return 0
	}
	return int64(math.Round(parsed))
}

func isVideoContentType(contentType string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	return strings.HasPrefix(contentType, "video/")
}
