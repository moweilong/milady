package initial

import (
	"context"
	"time"

	"github.com/moweilong/milady/pkg/logger"
	app "github.com/moweilong/milady/pkg/mapp"
	"github.com/moweilong/milady/pkg/tracer"

	"github.com/moweilong/milady/internal/config"
	"github.com/moweilong/milady/internal/database"
)

// Close releasing resources after service exit
func Close(servers []app.IServer) []app.Close {
	var closes []app.Close

	// close server
	for _, s := range servers {
		closes = append(closes, s.Stop)
	}

	// close database
	closes = append(closes, func() error {
		return database.CloseDB()
	})

	// close redis
	if config.Get().App.CacheType == "redis" {
		closes = append(closes, func() error {
			return database.CloseRedis()
		})
	}

	// close tracing
	if config.Get().App.EnableTrace {
		closes = append(closes, func() error {
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second) //nolint
			return tracer.Close(ctx)
		})
	}

	// close logger
	closes = append(closes, func() error {
		return logger.Sync()
	})

	return closes
}
