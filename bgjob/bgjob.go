package bgjob

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	localLog "github.com/daqnext/LocalLog/log"

	fj "github.com/daqnext/fastjson"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randJobId() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

const TYPE_PANIC_REDO = "panic_redo"
const TYPE_PANIC_RETURN = "panic_return"
const PANIC_REDO_SECS = 60

const STATUS_RUNNING string = "running"
const STATUS_WAITING string = "waiting"
const STATUS_CLOSING string = "closing"

type Job struct {
	jobName     string
	interval    int64
	createTime  int64
	lastRuntime int64
	info        *fj.FastJson
	status      string
	cycles      int64

	jobTYPE    string
	panic_redo chan struct{}
	panic_done chan struct{}

	context       interface{}
	processFn     func(interface{}, *fj.FastJson)
	chkContinueFn func(interface{}, *fj.FastJson) bool
	afCloseFn     func(interface{}, *fj.FastJson)
}

func (jb *Job) recordPanicStack(jm *JobManager, panicstr string, stack string) {

	errors := []string{panicstr}
	errstr := panicstr

	errors = append(errors, "last err unix-time:"+strconv.FormatInt(time.Now().Unix(), 10))

	lines := strings.Split(stack, "\n")
	maxlines := len(lines)
	if maxlines >= 100 {
		maxlines = 100
	}

	if maxlines >= 3 {
		for i := 2; i < maxlines; i = i + 2 {
			fomatstr := strings.ReplaceAll(lines[i], "	", "")
			errstr = errstr + "#" + fomatstr
			errors = append(errors, fomatstr)
		}
	}

	h := md5.New()
	h.Write([]byte(errstr))
	errhash := hex.EncodeToString(h.Sum(nil))

	jm.PanicExist = true
	jm.PanicJson.SetStringArray(errors, "errors", jb.jobName, errhash)

	jm.llog.Logger.Error("bgjob-catch-panic: ", " jobname:", jb.jobName, " errhash:", errhash, " errors:", errors)

}

type JobManager struct {
	AllJobs    sync.Map
	PanicExist bool
	PanicJson  *fj.FastJson
	llog       *localLog.LocalLog
}

func New(localLogger *localLog.LocalLog) *JobManager {
	fj.NewFromString("{}")
	return &JobManager{
		PanicExist: false,
		PanicJson:  fj.NewFromString("{}"),
		llog:       localLogger,
	}
}

func (jm *JobManager) ClearPanics() {
	jm.PanicExist = false
	jm.PanicJson = fj.NewFromString("{}")
}

func (jm *JobManager) StartJob_Panic_Redo(
	jobname string,
	interval int64,
	process_fn func(*fj.FastJson)) (string, error) {
	return jm.StartJobWithContext(TYPE_PANIC_REDO, jobname, interval, nil, func(i interface{}, fjh *fj.FastJson) {
		process_fn(fjh)
	}, nil, nil)
}

func (jm *JobManager) StartJob_Panic_Return(
	jobname string,
	interval int64,
	process_fn func(*fj.FastJson)) (string, error) {
	return jm.StartJobWithContext(TYPE_PANIC_RETURN, jobname, interval, nil, func(i interface{}, fjh *fj.FastJson) {
		process_fn(fjh)
	}, nil, nil)
}

func (jm *JobManager) StartJobWithContext(
	jobtype string,
	jobname string,
	interval int64,
	context interface{},
	process_fn func(interface{}, *fj.FastJson),
	chk_continue_fn func(interface{}, *fj.FastJson) bool,
	afclose_fn func(interface{}, *fj.FastJson)) (string, error) {

	if jobtype != TYPE_PANIC_REDO && jobtype != TYPE_PANIC_RETURN {
		return "", errors.New("job type error")
	}

	if interval < 1 {
		return "", errors.New("interval at least 1 second")
	}

	//generate a random job id that not exist yet
	jobid := ""
	for {
		jobid = randJobId()
		_, ok := jm.AllJobs.Load(jobid)
		if !ok {
			break
		}
	}

	createTime := time.Now().Unix()

	fjpointre := fj.NewFromString("{}")
	fjpointre.SetString(jobname, "jobName")
	fjpointre.SetString(STATUS_WAITING, "status")
	fjpointre.SetInt(0, "lastRuntime")
	fjpointre.SetInt64(createTime, "createTime")
	fjpointre.SetInt(0, "cycles")
	fjpointre.SetInt64(interval, "interval")
	fjpointre.SetString(jobtype, "jobTYPE")

	todoJobPointer := &Job{
		jobTYPE:       jobtype,
		jobName:       jobname,
		lastRuntime:   0,
		createTime:    createTime,
		status:        STATUS_WAITING,
		cycles:        0,
		interval:      interval,
		info:          fjpointre,
		context:       context,
		processFn:     process_fn,
		chkContinueFn: chk_continue_fn,
		afCloseFn:     afclose_fn,
		panic_redo:    make(chan struct{}),
		panic_done:    make(chan struct{}),
	}

	jm.AllJobs.Store(jobid, todoJobPointer)

	///start the monitoring routing
	go func(jobid_ string) {

		job, ok := jm.AllJobs.Load(jobid_)
		if !ok {
			return
		}
		for {

			select {
			case <-job.(*Job).panic_redo:
				go func(jobid_ string) {

					defer func() {
						if err := recover(); err != nil {
							//record panic
							var ErrStr string
							switch e := err.(type) {
							case string:
								ErrStr = e
							case runtime.Error:
								ErrStr = e.Error()
							case error:
								ErrStr = e.Error()
							default:
								ErrStr = "recovered (default) panic"
							}

							job.(*Job).recordPanicStack(jm, ErrStr, string(debug.Stack()))
							//check redo
							if job.(*Job).jobTYPE == TYPE_PANIC_REDO {
								time.Sleep(PANIC_REDO_SECS * time.Second)
								job.(*Job).panic_redo <- struct{}{}
							} else {
								job.(*Job).panic_done <- struct{}{}
							}
						}
					}()
					jm.dojob(jobid_)
				}(jobid_)
			case <-job.(*Job).panic_done:
				jm.AllJobs.Delete(jobid_)
				return
			}
		}
	}(jobid)

	todoJobPointer.panic_redo <- struct{}{}
	return jobid, nil
}

func (jm *JobManager) dojob(jobid_ string) {
	jobh_, ok := jm.AllJobs.Load(jobid_)
	if !ok {
		return
	}

	jobh := jobh_.(*Job)

	for {

		if ((jobh.chkContinueFn != nil) && (!jobh.chkContinueFn(jobh.context, jobh.info))) ||
			(jobh.status == STATUS_CLOSING) {

			jobh.status = STATUS_CLOSING
			jobh.info.SetString(STATUS_CLOSING, "Status")

			if jobh.afCloseFn != nil {
				jobh.afCloseFn(jobh.context, jobh.info)
			}
			jobh.panic_done <- struct{}{}
			return
		}

		nowUnixTime := time.Now().Unix()
		toSleepSecs := jobh.lastRuntime + jobh.interval - nowUnixTime
		if toSleepSecs <= 0 {
			jobh.lastRuntime = nowUnixTime
			jobh.info.SetInt64(jobh.lastRuntime, "LastRuntime")
			jobh.status = STATUS_RUNNING
			jobh.info.SetString(STATUS_RUNNING, "Status")
			jobh.cycles++
			jobh.info.SetInt64(jobh.cycles, "Cycles")
			// run
			jobh.processFn(jobh.context, jobh.info)
			//end
			jobh.status = STATUS_WAITING
			jobh.info.SetString(STATUS_WAITING, "Status")

		} else {
			time.Sleep(time.Duration(toSleepSecs) * time.Second)
		}
	}
}

//return nil if not exist
func (jm *JobManager) GetGBJob(jobid string) *Job {
	jobh_, ok := jm.AllJobs.Load(jobid)
	if ok {
		return jobh_.(*Job)
	} else {
		return nil
	}
}

func (jm *JobManager) CloseAndDeleteJob(jobid string) {
	jobh_, ok := jm.AllJobs.Load(jobid)
	if ok {
		jobh_.(*Job).status = STATUS_CLOSING
	}
}

func (jm *JobManager) CloseAndDeleteAllJobs() {
	jm.AllJobs.Range(func(_, value interface{}) bool {
		value.(*Job).status = STATUS_CLOSING
		return true
	})
}

func (jm *JobManager) GetAllJobsInfo() string {
	result := "["
	jm.AllJobs.Range(func(_, value interface{}) bool {
		result = result + value.(*Job).info.GetContentAsString()
		result = result + ","
		return true
	})
	result = strings.Trim(result, ",") + "]"
	return result
}
