package bgjob

import (
	"errors"
	"math/rand"
	"time"

	fj "github.com/daqnext/fastjson"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randJobId() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

const STATUS_RUNNING string = "running"
const STATUS_WAITING string = "waiting"
const STATUS_CLOSING string = "closing"

type Job struct {
	JobName     string
	Interval    int64
	CreateTime  int64
	LastRuntime int64
	Info        *fj.FastJson
	Status      string
	Cycles      int64

	Context       interface{}
	ProcessFn     func(interface{})
	ChkContinueFn func(interface{}) bool
	AfCloseFn     func(interface{})
}

var singleAllJobs map[string]*Job = make(map[string]*Job) //string is a hashtag

func StartJob(
	jobname string,
	interval int64,
	context interface{},
	process_fn func(interface{}),
	chk_continue_fn func(interface{}) bool,
	afclose_fn func(interface{})) (string, error) {

	if interval < 1 {
		return "", errors.New("interval at least 1 second")
	}

	//generate a random job id that not exist yet
	jobid := ""
	for {
		jobid = randJobId()
		_, ok := singleAllJobs[jobid]
		if !ok {
			break
		}
	}

	createTime := time.Now().Unix()

	fjpointre, _ := fj.NewFromString("{}")
	fjpointre.SetString(jobname, "JobName")
	fjpointre.SetString(STATUS_WAITING, "Status")
	fjpointre.SetInt(0, "LastRuntime")
	fjpointre.SetInt(createTime, "CreateTime")
	fjpointre.SetInt(0, "Cycles")
	fjpointre.SetInt(interval, "Interval")

	singleAllJobs[jobid] = &Job{
		JobName:       jobname,
		LastRuntime:   0,
		CreateTime:    createTime,
		Status:        STATUS_WAITING,
		Cycles:        0,
		Interval:      interval,
		Info:          fjpointre,
		Context:       context,
		ProcessFn:     process_fn,
		ChkContinueFn: chk_continue_fn,
		AfCloseFn:     afclose_fn,
	}

	go func(jobid_ string) {

		jobh := singleAllJobs[jobid_]
		for {

			if !jobh.ChkContinueFn(jobh.Context) || jobh.Status == STATUS_CLOSING {
				jobh.Status = STATUS_CLOSING
				jobh.Info.SetString(STATUS_CLOSING, "Status")
				jobh.AfCloseFn(jobh.Context)
				delete(singleAllJobs, jobid_)
				return
			}

			nowUnixTime := time.Now().Unix()
			toSleepSecs := jobh.LastRuntime + jobh.Interval - nowUnixTime
			if toSleepSecs <= 0 {
				jobh.LastRuntime = nowUnixTime
				jobh.Info.SetInt(jobh.LastRuntime, "LastRuntime")
				jobh.Status = STATUS_RUNNING
				jobh.Info.SetString(STATUS_RUNNING, "Status")
				jobh.Cycles++
				jobh.Info.SetInt(jobh.Cycles, "Cycles")
				// run
				jobh.ProcessFn(jobh.Context)
				//end
				jobh.Status = STATUS_WAITING
				jobh.Info.SetString(STATUS_WAITING, "Status")

			} else {
				time.Sleep(time.Duration(toSleepSecs) * time.Second)
			}
		}

	}(jobid)

	return jobid, nil
}

//return nil if not exist
func GetGBJob(jobid string) *Job {
	value, ok := singleAllJobs[jobid]
	if ok {
		return value
	} else {
		return nil
	}
}

func CloseAndDeleteJob(jobid string) {
	singleAllJobs[jobid].Status = STATUS_CLOSING
}

func CloseAndDeleteAllJobs() {
	for jobid := range singleAllJobs {
		singleAllJobs[jobid].Status = STATUS_CLOSING
	}
}
