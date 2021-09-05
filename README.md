# BGJOB_GO
background job util with go version

```
//install package:
go get github.com/daqnext/BGJOB_GO
```

```
// example:

 
	import 
	(
		"github.com/daqnext/BGJOB_GO/bgjob"
	)

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

```