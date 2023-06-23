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
			_groupIds := make([]int64, 0, len(groupIds))
			for id := range groupIds {
				_groupIds = append(_groupIds, id)
			}
			bot.ReportWeiboHot(_groupIds, true)

			privateIds := bot.state.reportState.getReportList(true)
			_privateIds := make([]int64, 0, len(privateIds))

			for id := range bot.state.reportState.getReportList(false) {
				_privateIds = append(_privateIds, id)
			}
			bot.ReportWeiboHot(_privateIds, false)
		},
		Corn:  corn,
		Model: top_list.Weibo,
	}
}

func (bot *CQBot) D36krHotReporter(corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			groupIds := bot.state.reportState.getReportList(true)
			_groupIds := make([]int64, 0, len(groupIds))
			for id := range groupIds {
				_groupIds = append(_groupIds, id)
			}
			bot.Report36kr(_groupIds, true)
			privateIds := bot.state.reportState.getReportList(true)
			_privateIds := make([]int64, 0, len(privateIds))

			for id := range bot.state.reportState.getReportList(false) {
				_privateIds = append(_privateIds, id)
			}

			bot.Report36kr(_privateIds, false)
		},
		Corn:  corn,
		Model: top_list.D36kr,
	}
}

func (bot *CQBot) WallStreetNewsReporter(corn string) *ReportJob {
	return &ReportJob{
		Report: func() {
			groupIds := bot.state.reportState.getReportList(true)
			_groupIds := make([]int64, 0, len(groupIds))
			for id := range groupIds {
				_groupIds = append(_groupIds, id)
			}

			bot.ReportWallStreetNews(_groupIds, true)

			privateIds := bot.state.reportState.getReportList(true)
			_privateIds := make([]int64, 0, len(privateIds))

			for id := range bot.state.reportState.getReportList(false) {
				_privateIds = append(_privateIds, id)
			}

			bot.ReportWallStreetNews(_privateIds, false)

		},
		Corn:  corn,
		Model: top_list.WallStreet,
	}
}

func init() {
	JobModels = map[string]string{}
}
