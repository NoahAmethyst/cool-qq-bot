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
	JobModels[r.Model] = r.Corn
	cron.AddCronJob(r.Report, r.Corn)
}

func (r *ReportJob) JobType() int {
	return Report
}

func (bot *CQBot) RegisterJob(job IJob) {
	if job == nil {
		log.Error("Register job failed:nil job")
		return
	}
	job.RunJob()
}

func (bot *CQBot) WeiboHotReporter(corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			groupIds := bot.state.reportState.getReportList(true)
			bot.ReportWeiboHot(groupIds, true)
			privateIds := bot.state.reportState.getReportList(false)
			bot.ReportWeiboHot(privateIds, false)
		},
		Corn:  corn,
		Model: top_list.Weibo,
	}
}

func (bot *CQBot) D36krHotReporter(corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			groupIds := bot.state.reportState.getReportList(true)
			bot.Report36kr(groupIds, true)
			privateIds := bot.state.reportState.getReportList(false)
			bot.Report36kr(privateIds, false)
		},
		Corn:  corn,
		Model: top_list.D36kr,
	}
}

func (bot *CQBot) WallStreetNewsReporter(corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			groupIds := bot.state.reportState.getReportList(true)
			bot.ReportWallStreetNews(groupIds, true)
			privateIds := bot.state.reportState.getReportList(false)
			bot.ReportWallStreetNews(privateIds, false)
		},
		Corn:  corn,
		Model: top_list.WallStreet,
	}
}

func (bot *CQBot) CaiXinNewsReporter(corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			groupIds := bot.state.reportState.getReportList(true)
			bot.ReportCaiXinNews(groupIds, true)
			privateIds := bot.state.reportState.getReportList(false)
			bot.ReportCaiXinNews(privateIds, false)
		},
		Corn:  corn,
		Model: top_list.WallStreet,
	}
}

func init() {
	JobModels = map[string]string{}
}
