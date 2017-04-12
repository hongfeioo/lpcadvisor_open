//  这里会生成一个可执行文件    uploadCadvisorData

package main

import (
	"os"
	"time"
)

func main() {

	// t1 := time.Now().Unix()
	// timestamp1 := fmt.Sprintf("%d", t1)
	timestamp1 := time.Now().Format("2006-01-02 15:04:05") //获取现在的时间，612345不是随意的数字
	LogRun("uploadCadvisorData.go:9   start->" + timestamp1)

	pushData()
	iAmAlive()

	// t2 := time.Now().Unix()
	// timestamp2 := fmt.Sprintf("%d", t2)
	timestamp2 := time.Now().Format("2006-01-02 15:04:05")

	LogRun("uploadCadvisorData.go:9   end->" + timestamp2)
}

func iAmAlive() {
	f, _ := os.OpenFile("test.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	defer f.Close()

	f.Write([]byte("b"))
}
