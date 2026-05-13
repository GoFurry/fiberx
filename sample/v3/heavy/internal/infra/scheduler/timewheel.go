package service

import (
	"fmt"
	"time"

	log "github.com/gofurry/fiberx/v3/heavy/internal/infra/logging"
	"github.com/rfyiamcool/go-timewheel"
)

var timeWheel *timewheel.TimeWheel

func InitTimeWheelOnStart() error {
	if err := StartTimeWheel(); err != nil {
		return err
	}
	log.Info("StartTimeWheel finish")
	return nil
}

func StartTimeWheel() error {
	if timeWheel != nil {
		return nil
	}

	tw, err := timewheel.NewTimeWheel(100*time.Millisecond, 1200, timewheel.TickSafeMode())
	if err != nil {
		return err
	}

	timeWheel = tw
	timeWheel.Start()
	return nil
}

func Stop() {
	if timeWheel != nil {
		timeWheel.Stop()
		timeWheel = nil
	}
}

func RemoveTask(task *timewheel.Task) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	log.Info(fmt.Sprintf("remove Cronn Job: %v", task))
	timeWheel.Remove(task)
}

func AddCronJob(tick time.Duration, job func()) *timewheel.Task {
	task := timeWheel.AddCron(tick, job)
	log.Info(fmt.Sprintf("AddOrUpdate Cron Job: %v", task))
	return task
}
