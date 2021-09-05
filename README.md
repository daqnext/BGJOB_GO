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
			fmt.Println("afterclose myjob1")
		})

	bgjob.StartJob("myjob2", 2, nil,
		func(c interface{}) {
			fmt.Println("proccessing myjob2")
		}, func(c interface{}) bool {
			return true
		}, func(c interface{}) {
			fmt.Println("afterclose myjob2")
		})
```