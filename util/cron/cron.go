package cron

import (
	"github.com/robfig/cron"
)


const (
	JobDurationHourly   = "@hourly"
	JobDurationDaily    = "@daily"
	JobDurationMinutely = "@every 60s"
	JobDuration10S      = "@every 10s"
)

// AddCronJob 时任务执行器
func AddCronJob(cronJob func(), jobDuration string) {

	c := cron.New()
	_ = c.AddFunc(jobDuration, func() {
		cronJob()
	})
	c.Start()
}
