package main

import (
	"fmt"
	"time"

	"github.com/daqnext/BGJOB_GO/bgjob"
)

func main() {

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
			bgjob.CloseAndDeleteAllJobs()
		})

	bgjob.StartJob("myjob2", 2, nil,
		func(c interface{}) {
			fmt.Println("proccessing myjob2")
		}, func(c interface{}) bool {
			return true
		}, func(c interface{}) {
			fmt.Println("afterclose myjob2")
		})

	time.Sleep(400 * time.Second)
}
