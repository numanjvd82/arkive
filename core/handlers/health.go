package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/database"
)

func Health(db database.PgExecutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.QueryRow(ctx, "select 1").Scan(new(int)); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "db_unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
