package main

import (
	"fmt"
	"time"

	"github.com/daqnext/BGJOB_GO/bgjob"
	fj "github.com/daqnext/fastjson"
)

func main() {

	bgmh := bgjob.New()

	type mycontext struct {
		Counter int
	}

	bgmh.StartJob("myjob1", 2, &mycontext{Counter: 0},
		func(c interface{}, fjh *fj.FastJson) {
			fmt.Println("proccessing")
			c.(*mycontext).Counter++
			fjh.SetInt(int64(c.(*mycontext).Counter), "Counter")

		}, func(c interface{}, fjh *fj.FastJson) bool {
			fmt.Println(fjh.GetContentAsString())
			if c.(*mycontext).Counter == 5 {
				return false
			}
			return true
		}, func(c interface{}, fjh *fj.FastJson) {
			fmt.Println("afterclose")
			fmt.Println("will close all jobs")
		})

	bgmh.StartJob("myjob2", 2, nil, func(c interface{}, fjh *fj.FastJson) {
		fmt.Println("proccessing myjob2")
	}, nil, nil)

	fmt.Println("///////////////////////")
	fmt.Println(bgmh.GetAllJobsInfo())
	fmt.Println("///////////////////////")

	time.Sleep(400 * time.Second)
}
