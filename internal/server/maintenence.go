package server

import (
	"context"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Maintenance interface {
	ID() string
	Clean(ctx context.Context) error
}

type GroupMaintenance struct {
	logger          *zap.Logger
	maintenanceList []Maintenance
}

func NewGroupMaintenance(logger *zap.Logger, m ...Maintenance) GroupMaintenance {
	return GroupMaintenance{
		logger:          logger,
		maintenanceList: m,
	}
}

func (o *GroupMaintenance) Start(ctx context.Context, interval time.Duration) {
	var (
		ticker = time.NewTicker(interval)
		wg     sync.WaitGroup
		index  = uint64(0)
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		wg.Add(len(o.maintenanceList))

		for i := 0; i < len(o.maintenanceList); i++ {
			m := o.maintenanceList[i]

			go func() {
				defer wg.Done()

				o.logger.Info("clean: start",
					zap.String("id", m.ID()),
				)

				start := time.Now()

				err := m.Clean(ctx)
				if err != nil {
					o.logger.Error("failed to clean",
						zap.String("id", m.ID()),
						zap.Error(err),
					)

					return
				}

				o.logger.Info("clean: finish",
					zap.Duration("duration", time.Since(start)),
					zap.String("id", m.ID()),
				)
			}()
		}

		wg.Wait()

		index++
		if index%10 == 0 {
			go runtime.GC()
		}
	}
}
