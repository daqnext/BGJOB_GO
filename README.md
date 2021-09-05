# BGJOB_GO
background job util with go version

```go
//install package:
go get github.com/daqnext/BGJOB_GO

import 
(
	"github.com/daqnext/BGJOB_GO/bgjob"
)
```

```go
// example:

	//////////////	/////////////////	
	// chk_continue_fn is invoked before each cycle 
	// process_fn [the main function ] will be called if chk_continue_fn return true
	// afclose_fn will be called if chk_continue_fn return false

	func StartJob(
	jobname string,
	interval int64,
	context interface{},
	process_fn func(interface{}),
	chk_continue_fn func(interface{}) bool,
	afclose_fn func(interface{})) (string, error)


	/// use case 1  job with context ////////////
	type mycontext struct {
			Counter int
	}

	bgjob.StartJob("myjob1", 2, &mycontext{Counter: 0},
		func(c interface{}) {
			fmt.Println("proccessing")
			c.(*mycontext).Counter++
		}, func(c interface{}) bool {
			fmt.Println("checking:", c.(*mycontext).Counter)
			if c.(*mycontext).Counter == 5 {
				return false
			}
			return true
		}, func(c interface{}) {
			fmt.Println("afterclose")
			fmt.Println("will close all jobs")
			//bgjob.CloseAndDeleteAllJobs()
		})

	////// use case 2 ////////////
	var jid string
	jid, _ = bgjob.StartJob("myjob2", 5, nil,
		func(c interface{}) {
			fmt.Println(bgjob.GetGBJob(jid).Info.GetContentAsString())
		}, func(c interface{}) bool {
			fmt.Println(bgjob.GetGBJob(jid).Info.GetContentAsString())
			return true
		}, func(c interface{}) {
			//fmt.Println("afterclose myjob2")
		})

	// get the job with job id 
	bgjob.GetGBJob(jid)

	// get the log of the job 
	bgjob.GetGBJob(jid).Info.GetContentAsString()


	//time.Sleep(400 * time.Second)
 
```