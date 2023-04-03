package coolq

import (
	"github.com/Mrs4s/go-cqhttp/util/cron"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	log "github.com/sirupsen/logrus"
)

const (
	Report = iota
)

var JobModels map[string]string

type (
	IJob interface {
		RunJob()
		JobType() int
	}

	ReportJob struct {
		Report func()
		Model  string
		Corn   string
	}
)

func (r *ReportJob) RunJob() {
	cron.AddCronJob(r.Report, r.Corn)
}

func (r *ReportJob) JobType() int {
	return Report
}

func (bot *CQBot) RegisterJob(job IJob) {
	switch job.JobType() {
	case Report:
		reportJob := job.(*ReportJob)
		reportJob.RunJob()
		JobModels[reportJob.Model] = reportJob.Corn
		log.Infof("注册定时推送任务：%s", reportJob.Model)
	default:

	}
}

func (bot *CQBot) WeiboHotReporter(group int64, corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			bot.ReportWeiboHot(group)
		},
		Corn:  corn,
		Model: top_list.Weibo,
	}
}

func (bot *CQBot) D36krHotReporter(group int64, corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			bot.Report36kr(group)
		},
		Corn:  corn,
		Model: top_list.D36kr,
	}
}

func (bot *CQBot) WallStreetNewsReporter(group int64, corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			bot.ReportWallStreetNews(group)
		},
		Corn:  corn,
		Model: top_list.WallStreet,
	}
}

func init() {
	JobModels = map[string]string{}
}
