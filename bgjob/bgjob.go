package bgjob

import (
	"errors"
	"math/rand"
	"strings"
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
	ProcessFn     func(interface{}, *fj.FastJson)
	ChkContinueFn func(interface{}, *fj.FastJson) bool
	AfCloseFn     func(interface{}, *fj.FastJson)
}

type JobManager struct {
	AllJobs map[string]*Job
}

func New() *JobManager {
	return &JobManager{AllJobs: make(map[string]*Job)}
}

func (jm *JobManager) StartJob(
	jobname string,
	interval int64,
	process_fn func(*fj.FastJson)) (string, error) {
	return jm.StartJobWithContext(jobname, interval, nil, func(i interface{}, fjh *fj.FastJson) {
		process_fn(fjh)
	}, nil, nil)
}

func (jm *JobManager) StartJobWithContext(
	jobname string,
	interval int64,
	context interface{},
	process_fn func(interface{}, *fj.FastJson),
	chk_continue_fn func(interface{}, *fj.FastJson) bool,
	afclose_fn func(interface{}, *fj.FastJson)) (string, error) {

	if interval < 1 {
		return "", errors.New("interval at least 1 second")
	}

	//generate a random job id that not exist yet
	jobid := ""
	for {
		jobid = randJobId()
		_, ok := jm.AllJobs[jobid]
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

	jm.AllJobs[jobid] = &Job{
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

		jobh := jm.AllJobs[jobid_]
		for {

			if ((jobh.ChkContinueFn != nil) && (!jobh.ChkContinueFn(jobh.Context, jobh.Info))) ||
				(jobh.Status == STATUS_CLOSING) {

				jobh.Status = STATUS_CLOSING
				jobh.Info.SetString(STATUS_CLOSING, "Status")

				if jobh.AfCloseFn != nil {
					jobh.AfCloseFn(jobh.Context, jobh.Info)
				}
				delete(jm.AllJobs, jobid_)
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
				jobh.ProcessFn(jobh.Context, jobh.Info)
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
func (jm *JobManager) GetGBJob(jobid string) *Job {
	value, ok := jm.AllJobs[jobid]
	if ok {
		return value
	} else {
		return nil
	}
}

func (jm *JobManager) CloseAndDeleteJob(jobid string) {
	jm.AllJobs[jobid].Status = STATUS_CLOSING
}

func (jm *JobManager) CloseAndDeleteAllJobs() {
	for jobid := range jm.AllJobs {
		jm.AllJobs[jobid].Status = STATUS_CLOSING
	}
}

func (jm *JobManager) GetAllJobsInfo() string {
	result := "["
	for jobid := range jm.AllJobs {
		result = result + jm.AllJobs[jobid].Info.GetContentAsString()
		result = result + ","
	}
	result = strings.Trim(result, ",") + "]"
	return result
}
