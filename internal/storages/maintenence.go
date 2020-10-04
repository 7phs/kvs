package storages

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Maintenance interface {
	Clean(ctx context.Context) error
}

type GroupMaintenance struct {
	logger          *zap.Logger
	maintenanceList []Maintenance
}

func NewGroupMaintenance(logger *zap.Logger, m ...Maintenance) *GroupMaintenance {
	return &GroupMaintenance{
		logger:          logger,
		maintenanceList: m,
	}
}

func (o *GroupMaintenance) Start(ctx context.Context, interval time.Duration) {
	var (
		ticker = time.NewTicker(interval)
		wg     sync.WaitGroup
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
				err := m.Clean(ctx)
				if err != nil {
					o.logger.Error("failed to maintenance",
						zap.Error(err),
					)
				}
			}()
		}

		wg.Wait()
	}
}
