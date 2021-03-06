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
