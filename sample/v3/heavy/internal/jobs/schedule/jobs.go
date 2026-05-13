package schedule

import (
	"time"

	env "github.com/gofurry/fiberx/v3/heavy/config"
	"github.com/gofurry/fiberx/v3/heavy/internal/jobs/schedule/task"
)

type Job struct {
	Name       string
	Interval   time.Duration
	RunOnStart bool
	Run        func()
}

func Jobs() []Job {
	return []Job{
		{
			Name:       "schedule.metrics_cache_refresh",
			Interval:   10 * time.Minute,
			RunOnStart: true,
			Run:        ScheduleByTenMinutes,
		},
		{
			Name:       "schedule.hourly_housekeeping",
			Interval:   1 * time.Hour,
			RunOnStart: true,
			Run:        ScheduleByOneHour,
		},
	}
}

func ScheduleByTenMinutes() {
	cfg := env.GetServerConfig()

	if cfg.Prometheus.Enabled {
		task.UpdateMetricsCache()
	}
}

func ScheduleByOneHour() {
}
