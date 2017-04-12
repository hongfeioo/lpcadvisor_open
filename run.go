//  这里会生成一个可执行文件 run ， 也是micadvistor 容器启动后运行的第一个可执行文件。
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

//Interval 检测时间间隔
var Interval time.Duration

func main() {
	tmp := os.Getenv("Interval")
	Interval = 60 * time.Second
	tmp1, err := strconv.ParseInt(tmp, 10, 64)
	//fmt.Println(tmp1)
	if err != nil {
		LogErr(err, "run.go:21  strconv.ParseInt error!")
		return
	}
	Interval = time.Duration(tmp1) * time.Second

	cmd := exec.Command("/home/work/uploadCadviosrData/cadvisor")
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		LogErr(err, "run.go:32  exec  cadvisor error!")
		return
	}

	LogRun("run.go  start cadvisor successful")

	go func() {
		t := time.NewTicker(Interval)
		for {
			<-t.C
			LogRun("run.go  start new goroutine  uploadCadvisorData")
			cmd = exec.Command("/home/work/uploadCadviosrData/uploadCadvisorData") //undefined: exec.CommandContext   这里需要使用context控制程序的结束时间
			if err := cmd.Start(); err != nil {
				fmt.Println(err)
				LogErr(err, "run.go  cmd.start  uploadCadvisorData error!")
				return
			}
			if err := cmd.Wait(); err != nil {
				LogErr(err, "run.go  cmd.wait   uploadCadvisorData error!")

			}

		}
	}()

	for {
		time.Sleep(time.Second * 240)
		if isAlive() {
			clean()
		} else {
			LogErr(nil, "run.go:57  check uploadCadvisroData fail")
			os.Exit(1)
		}
	}

}
func isAlive() bool {
	f, _ := os.OpenFile("test.txt", os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0660)
	defer f.Close()
	readBuf := make([]byte, 32)
	var pos int64
	n, _ := f.ReadAt(readBuf, pos)
	if n == 0 {
		return false
	}
	return true
}

func clean() {
	f, _ := os.OpenFile("test.txt", os.O_TRUNC, 0660)
	defer f.Close()
}
