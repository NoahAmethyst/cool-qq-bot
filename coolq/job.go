package coolq

import (
	"github.com/Mrs4s/go-cqhttp/util/cron"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
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
	JobModels[r.Model] = r.Corn
	cron.AddCronJob(r.Report, r.Corn)
}

func (r *ReportJob) JobType() int {
	return Report
}

func (bot *CQBot) RegisterJob(job IJob) {
	job.RunJob()
}

func (bot *CQBot) WeiboHotReporter(group int64, corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			bot.ReportWeiboHot(group, true)
		},
		Corn:  corn,
		Model: top_list.Weibo,
	}
}

func (bot *CQBot) D36krHotReporter(group int64, corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			bot.Report36kr(group, true)
		},
		Corn:  corn,
		Model: top_list.D36kr,
	}
}

func (bot *CQBot) WallStreetNewsReporter(group int64, corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			bot.ReportWallStreetNews(group, true)
		},
		Corn:  corn,
		Model: top_list.WallStreet,
	}
}

func init() {
	JobModels = map[string]string{}
}
