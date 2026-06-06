package jobs

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"

	filessvc "arkive/core/services/files"
)

const deletedFileCleanupSchedule = "@every 2h"

func StartDeletedFileCleanup(filesService *filessvc.Service) (*cron.Cron, error) {
	logger := log.New(os.Stdout, "cron ", log.LstdFlags)
	cleanupCron := cron.New(
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
		cron.WithLogger(cron.VerbosePrintfLogger(logger)),
	)
	_, err := cleanupCron.AddFunc(deletedFileCleanupSchedule, func() {
		// TODO: add a DB advisory lock before running cleanup when scaling to multiple instances.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		count, cleanupErr := filesService.PurgeDeletedFiles(ctx)
		if cleanupErr != nil {
			log.Printf("deleted file cleanup failed: %v", cleanupErr)
			return
		}
		log.Printf("deleted file cleanup completed: %d purged files", count)
	})
	if err != nil {
		return nil, err
	}
	cleanupCron.Start()
	return cleanupCron, nil
}
