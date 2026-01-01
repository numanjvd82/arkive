package jobs

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"

	"arkive/core/services/uploads"
)

const uploadCleanupSchedule = "@every 15m"

func StartUploadCleanup(uploadService *uploads.Service) (*cron.Cron, error) {
	logger := log.New(os.Stdout, "cron ", log.LstdFlags)
	cleanupCron := cron.New(
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
		cron.WithLogger(cron.VerbosePrintfLogger(logger)),
	)
	_, err := cleanupCron.AddFunc(uploadCleanupSchedule, func() {
		// TODO: add a DB advisory lock before running cleanup when scaling to multiple instances.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		count, cleanupErr := uploadService.CleanupExpiredUploads(ctx)
		if cleanupErr != nil {
			log.Printf("upload cleanup failed: %v", cleanupErr)
			return
		}
		log.Printf("upload cleanup completed: %d expired uploads", count)
	})
	if err != nil {
		return nil, err
	}
	cleanupCron.Start()
	return cleanupCron, nil
}
