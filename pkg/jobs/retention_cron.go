package jobs

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"

	"arkive/core/services/uploads"
)

const retentionSchedule = "@every 6h"

func StartRetentionCleanup(uploadService *uploads.Service) (*cron.Cron, error) {
	logger := log.New(os.Stdout, "cron ", log.LstdFlags)
	retentionCron := cron.New(
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
		cron.WithLogger(cron.VerbosePrintfLogger(logger)),
	)
	_, err := retentionCron.AddFunc(retentionSchedule, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		marked, deleted, cleanupErr := uploadService.ApplyInactivityRetention(ctx)
		if cleanupErr != nil {
			log.Printf("retention cleanup failed: %v", cleanupErr)
			return
		}
		log.Printf("retention cleanup completed: %d marked, %d deleted", marked, deleted)
	})
	if err != nil {
		return nil, err
	}
	retentionCron.Start()
	return retentionCron, nil
}
