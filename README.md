# BGJOB_GO
background job util with go version

### background job can be rebooted using Panic_Redo type

### !!important : Don't write your own go routine inside background job 
### if you really want to write go routine inside bg job please use safe go-routine tool below:
### https://github.com/daqnext/go-smart-routine


```go
//install package:
go get github.com/daqnext/BGJOB_GO

import 
(
	"github.com/daqnext/BGJOB_GO/bgjob"
)
```

```go
	///////////////////////////////	
	// chk_continue_fn is invoked before each cycle 
	// process_fn [the main function ] will be called if chk_continue_fn return true
	// afclose_fn will be called if chk_continue_fn return false

	func StartJob(
	jobname string,
	interval int64,
	context interface{},
	process_fn func(interface{},fjh *fj.FastJson),
	chk_continue_fn func(interface{},fjh *fj.FastJson) bool,
	afclose_fn func(interface{},fjh *fj.FastJson)) (string, error)
```

```go
// example:
/// use case 1  job with context ////////////
 

package main

import (
	"fmt"
	"time"

	"github.com/daqnext/BGJOB_GO/bgjob"
	fj "github.com/daqnext/fastjson"

	localLog "github.com/daqnext/LocalLog/log"
)

func divide(a, b int) int {
	return a / b
}

func main() {

	lg, err := localLog.New("logs", 10, 10, 10)
	if err != nil {
		panic(err)
	}

	bgmh := bgjob.New(lg)

	type mycontext struct {
		Counter int
	}

	/////example of panic recovery background job
	x := 0
	bgmh.StartJob_Panic_Redo("myjob2", 2, func(fjh *fj.FastJson) {
		fjh.SetBoolean(true, "job2started")
		fmt.Println("proccessing myjob2")

		fmt.Println("start of the program")
		if x == 0 {
			x++
			fmt.Println("here0")
			divide(10, 0)
		}
		fmt.Println("end of the program")

	})

	bgmh.StartJobWithContext(bgjob.TYPE_PANIC_RETURN, "myjob1", 2, &mycontext{Counter: 0},
		func(c interface{}, fjh *fj.FastJson) {
			fmt.Println("myjob1 start proccessing")
			c.(*mycontext).Counter++
			fjh.SetInt64(int64(c.(*mycontext).Counter), "Counter")

		}, func(c interface{}, fjh *fj.FastJson) bool {
			fmt.Println(fjh.GetContentAsString())
			if c.(*mycontext).Counter == 5 {
				return false
			}
			return true
		}, func(c interface{}, fjh *fj.FastJson) {
			fmt.Println("myjob1 afterclose ")
		})

	time.Sleep(65 * time.Second)
	fmt.Println("///////////////////////")
	fmt.Println(bgmh.GetAllJobsInfo())
	fmt.Println("///////////////////////")
	if bgmh.PanicExist {
		fmt.Println("errors:", bgmh.PanicJson.GetContentAsString())
		bgmh.ClearPanics()
	}
	time.Sleep(400 * time.Second)
}


 
```