//main  uploadCadvisorData 程序会调用这里的函数，这里控制着获取并推送数据的主节奏

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	cpuNum   int64
	countNum int
)

func pushData() {
	//AllDockerBrief 变量存放所有docker的基本信息JSON格式 ，   原理  (echo -e "GET /containers/json HTTP/1.1\r\nHost: www.test.com\r\n")|nc -U //var/run/docker.sock
	AllDockerBrief, err1 := getAllDockerBrief()
	if err1 != nil {
		LogErr(err1, "from pushDatas.go getAllDockerBrief. return")
		return
	}

	var a []AllDockerBriefStruct
	if err := json.Unmarshal([]byte(AllDockerBrief), &a); err != nil {
		LogErr(err, "from pushDatas.go unmarshal AllDockerBrief. return  ")
		return
	}

	//循环读出所有的容器ID, 以容器ID为唯一标示去取所有内容。
	for i := 0; i < len(a); i++ {
		LogRun("=========>Begin [" + strconv.Itoa(i) + "/" + strconv.Itoa(len(a)) + "] ->" + a[i].ID + "->" + a[i].Names[0] + "=========>")
		//-----------------------------------
		//--- 通过容器Id获取inspect信息------   原理：(echo -e "GET /containers/XXXXXXXXXX/json HTTP/1.1\r\nHost: www.test.com\r\n")|nc -U //var/run/docker.sock
		DockerInspect, err := getDockerData(a[i].ID)
		if err != nil {
			LogErr(err, "from pushDatas.go getDockerData ..continue ")
			continue
		}
		var b DockerInspectStruct
		if err := json.Unmarshal([]byte(DockerInspect), &b); err != nil {
			LogErr(err, "from pushDatas.go unmarshal DockerInspect ..continue")
			continue
		}
		// inspect结构体中信息非常丰富，目前只取出了很少的信息，以后可以多加利用 ,例如 Env ，Labels等
		endpoint := getEndPoint(b.Name) // 容器的名字也可以在AllDockerBrief结构体中获取
		containerIP := b.NetworkSettings.IPAddress + "/" + strconv.Itoa(b.NetworkSettings.IPPrefixLen)
		tag := "" //函数已经有了，目前没用上，对应的容器的label

		//LogRun("=>" + endpoint + "=>" + containerIP) //  可以通过inspect获取容器的IP

		//----------------------------------------
		//---通过容器ID, 获取容器对应的cadvisor的详细信息---
		OneContainerCadvisorJSON, err2 := getOneContainerCadvisorData(a[i].ID)
		if err2 != nil {
			LogErr(err2, "from pushDatas.go getOneContainerCadvisorData  .return  ")
			return
		}

		CutIndex := strings.IndexAny(OneContainerCadvisorJSON, ":")
		if CutIndex == -1 {
			LogErr(nil, "from pushDatas.go CutIndex,can not find maohao  .return  ")
			return
		}
		// 由于原始的JSON串前边的key是不固定的字符，无法进行json解析，所以需要先截取成可以解析的结构体JSON
		OneContainerCadvisorJSONAfterCut := OneContainerCadvisorJSON[CutIndex+1 : len(OneContainerCadvisorJSON)-1]

		var c ContainerInfo // 使用官方cadvisor api v1中提供的api进行解析
		if err := json.Unmarshal([]byte(OneContainerCadvisorJSONAfterCut), &c); err != nil {
			LogErr(err, "from pushDatas.go unmarshal  ContainerInfo  ..return")
			continue
		}

		//----------------------------------------
		//--下边就可以任意取出cadvisor对容器的描述了---
		//LogRun("=>" + "开始获取这个容器的cadvisor数据" + c.ContainerReference.Id)
		timestamp := fmt.Sprintf("%d", time.Now().Unix()) //时间格式变化会上报失败，timestamp := time.Now().Format("2006-01-02 15:04:05")

		//---内存上报---------
		if err := pushMem(c, timestamp, tag, a[i].ID, endpoint+"-"+containerIP); err != nil { //get cadvisor data about Memery
			LogErr(err, "from pushDatas.go pushMem function  ")
		}

		//-----cpu上报------------
		if err := NewPushCPU(c, timestamp, tag, a[i].ID, endpoint+"-"+containerIP); err != nil {
			LogErr(err, "from pushDatas.go NewPushCPU function  ")

		}

		//-----IO上报-------------
		if err := pushDiskIo(c, timestamp, tag, a[i].ID, endpoint); err != nil {
			LogErr(err, "from pushDatas.go pushDisoIo function")
		}

		//------网卡流量上报--------
		if err := NewPushNetwork(c, timestamp, tag, a[i].ID, endpoint+"-"+containerIP); err != nil {
			LogErr(err, "from pushDatas.go NewPushNetwork function  ")
		}

	}

	// os.Exit(1)
}

// 本函数上报了四个数值，内存使用量，内存最大值，内存百分比， 如果需要上报的更多， 可以增加这个函数的参数。
func pushMem(_c ContainerInfo, timestamp, tags, containerID, endpoint string) error {
	LogRun("begin to push Mem Info")

	memLimit := _c.Spec.Memory.Limit //资源配额是个固定值，取一次即可

	var memUseage uint64
	for _, u := range _c.Stats {
		//LogRun(strconv.Itoa(i) + fmt.Sprint(u.Memory.Usage))
		memUseage = memUseage + u.Memory.Usage
	}
	memUseage = (memUseage) / uint64(len(_c.Stats)) // 内存用量是个变动数值，把stats数组中的所有数值加起来，求平均，以防止某一个为空造成的取值不准。

	//LogRun(fmt.Sprint(float64(memUseage)) + "---" + fmt.Sprint(float64(memLimit)))

	//上报内存使用率
	memUsagePrecent := float64(memUseage) / float64(memLimit) // 注意上报的这个数值已经乘以了100，即：上报的是百分数
	if err := pushIt(fmt.Sprintf("%.1f", memUsagePrecent*100), timestamp, "mem.memused.percent", tags, containerID, "GAUGE", endpoint); err != nil {
		// LogErr(err, "pushIt err in pushMem")
		return err
	}
	// 上报内存使用量
	if err := pushIt(fmt.Sprint(memUseage), timestamp, "mem.memused", tags, containerID, "GAUGE", endpoint); err != nil {
		//LogErr(err, "pushIt err in pushMem")
		return err
	}
	// 上报内存总量
	if err := pushIt(fmt.Sprint(memLimit), timestamp, "mem.memtotal", tags, containerID, "GAUGE", endpoint); err != nil {
		//LogErr(err, "pushIt err in pushMem")
		return err
	}

	return nil

}

// pushDiskIo  用于上报磁盘使用情况，
func pushDiskIo(_c ContainerInfo, timestamp, tags, containerID, endpoint string) error {
	LogRun("begin to push DiskIo Info")
	var writeUsage uint64
	var readUsage uint64
	var StatsCount uint64 //计数器存放， read 或者 write 值被累加的次数，作为求平均的分母；

	for _, u := range _c.Stats {
		StatsCount++

		for _, j := range u.DiskIo.IoServiceBytes {
			//这个数组的长度是4，不知道是否为固定的，不过无所谓， 全部进行和运算。
			//LogRun(strconv.Itoa(i) + strconv.Itoa(k) + "--" + fmt.Sprint(j.Stats["Read"]) + "---" + fmt.Sprint(j.Stats["Write"]))
			writeUsage += j.Stats["Write"]
			readUsage += j.Stats["Read"]
		}
	}

	writeUsage = writeUsage / StatsCount
	readUsage = readUsage / StatsCount

	if err := pushIt(fmt.Sprintf("%.2f", float64(writeUsage)/1048576), timestamp, "disk.io.write_MBps", tags, containerID, "GAUGE", endpoint); err != nil {
		LogErr(err, "pushIt err in pushDiskIo")
	}
	if err := pushIt(fmt.Sprintf("%.2f", float64(readUsage)/1048576), timestamp, "disk.io.read_MBps", tags, containerID, "GAUGE", endpoint); err != nil {
		LogErr(err, "pushIt err in pushDiskIo")
	}

	return nil
}

//NewPushCPU  函数用于上报stats结构体中cpu的信息，这些是全部的信息了。
//简单的做了一个除法，从ns->us->ms->s换算为s , 这个数值应该可以说明cpu的事情情况，以后有更优秀的算法再改进。
func NewPushCPU(_c ContainerInfo, timestamp, tags, containerID, endpoint string) error {
	LogRun("begin to push CPU Info")

	// 以下三个参数是可以提供上报的，暂时隐藏掉。
	// cpuUsageUser := _c.Stats[STATSINDEX].Cpu.Usage.User        用户空间占用的cpu量
	// cpuUsageSystem := _c.Stats[STATSINDEX].Cpu.Usage.System     内核空间占用的cpu值
	// cpuUsagePerCPUUsage := _c.Stats[STATSINDEX].Cpu.Usage.PerCpu     // 每个cpu核心占用的cpu量值

	var cpuLoadAverage int32
	var cpuUsageTotal uint64
	for _, u := range _c.Stats {

		//LogRun(strconv.Itoa(i) + "--" + fmt.Sprint(u.Cpu.LoadAverage) + "---" + fmt.Sprint(u.Cpu.Usage.Total))
		cpuLoadAverage += u.Cpu.LoadAverage
		cpuUsageTotal += u.Cpu.Usage.Total

	}
	cpuLoadAverage = cpuLoadAverage / int32(len(_c.Stats))
	cpuUsageTotal = cpuUsageTotal / uint64(len(_c.Stats))

	// 上报cpuload平均值， 这个值测试阶段一直为0，不知道线上知否能用到
	if err := pushIt(fmt.Sprintf("%.2f", float64(cpuLoadAverage)), timestamp, "cpu.loadaverage", tags, containerID, "GAUGE", endpoint); err != nil {
		return err
	}
	//上报cpu的总使用量，包括用户空间＋内核  1000,000,000,000,000 * 100%   这个数值和docker stats 观察到的cpu百分比很像， 估且就当作百分比吧，如果不是至少可以从这个数值看出cpu的负载情况。
	//即(cpu耗时 / 百万分之一秒) *100%   和cpu使用率很像。
	if err := pushIt(fmt.Sprintf("%.5f", float64(cpuUsageTotal)/10000000000000), timestamp, "cpu.usageTotalPercent", tags, containerID, "GAUGE", endpoint); err != nil {
		return err
	}
	// if err := pushIt(fmt.Sprintf("%.2f", float64(cpuUsageUser)/1000000000), timestamp, "cpu.UsageUser", tags, containerID, "GAUGE", endpoint); err != nil {
	// 	return err
	// }
	// if err := pushIt(fmt.Sprintf("%.2f", float64(cpuUsageSystem)/1000000000), timestamp, "cpu.UsageSystem", tags, containerID, "GAUGE", endpoint); err != nil {
	// 	return err
	// }
	// for i, u := range cpuUsagePerCPUUsage {
	//
	// 	if err := pushIt(fmt.Sprintf("%.2f", float64(u)/1000000000), timestamp, "cpu.perCpuUsage-"+strconv.Itoa(i), tags, containerID, "GAUGE", endpoint); err != nil {
	// 		return err
	// 	}
	//
	// }
	return nil

}

//NewPushNetwork 函数作用为容器上报网卡流量
// 由于采集的是网卡的累计值，每次采集的stats数组数量又不确定，所以书写算法：  stats_rxbytes[i+1] - stats_rxbytes[i]   , 依次可以计算出多个增量值， 返回增量值的平均数。
// 流量的单位是每秒， 丢包是以每秒平均丢包个数计算的
func NewPushNetwork(_c ContainerInfo, timestamp, tags, containerID, endpoint string) error {
	LogRun("begin to push NetWork Info")

	var rxBytes uint64
	var rxDropped uint64
	var txBytes uint64
	var txDropped uint64

	LogRun("NeWPushNetwork len([]Stats) should be >= 2 , now is " + strconv.Itoa(len(_c.Stats)))

	// 这种情况说明Stats数组太短，无法看出趋势
	if len(_c.Stats) < 2 {

		//push 所有数值为0 ，防止图形断点
		if err := pushIt("0", timestamp, "net.rx.KBps", tags, containerID, "GAUGE", endpoint); err != nil {
			return err
		}
		if err := pushIt("0", timestamp, "net.tx.KBps", tags, containerID, "GAUGE", endpoint); err != nil {
			return err
		}
		if err := pushIt("0", timestamp, "net.rxDrop.packets", tags, containerID, "GAUGE", endpoint); err != nil {
			return err
		}
		if err := pushIt("0", timestamp, "net.txDrop.packets", tags, containerID, "GAUGE", endpoint); err != nil {
			return err
		}

		return nil

	}

	// 控制循环的次数 ＝ 数组长度－1
	for i := 0; i < len(_c.Stats)-1; i++ {
		// 调试错误的时候可以放开
		//LogRun(strconv.Itoa(i) + "--" + fmt.Sprint(_c.Stats[i+1].Network.RxBytes-_c.Stats[i].Network.RxBytes) + "---" + fmt.Sprint(_c.Stats[i+1].Network.TxBytes-_c.Stats[i].Network.TxBytes))
		//LogRun(strconv.Itoa(i) + "--" + fmt.Sprint(_c.Stats[i+1].Network.RxDropped-_c.Stats[i].Network.RxDropped) + "---" + fmt.Sprint(_c.Stats[i+1].Network.TxDropped-_c.Stats[i].Network.TxDropped))

		rxBytes += _c.Stats[i+1].Network.RxBytes - _c.Stats[i].Network.RxBytes
		txBytes += _c.Stats[i+1].Network.TxBytes - _c.Stats[i].Network.TxBytes
		rxDropped += _c.Stats[i+1].Network.RxDropped - _c.Stats[i].Network.RxDropped
		txDropped += _c.Stats[i+1].Network.TxDropped - _c.Stats[i].Network.TxDropped

	}

	rxBytes = rxBytes / uint64(len(_c.Stats)-1) //本次统计周期的tx增量平均值
	txBytes = txBytes / uint64(len(_c.Stats)-1) //本次统计周期的tx增量平均值
	rxDropped = rxDropped / uint64(len(_c.Stats)-1)
	txDropped = txDropped / uint64(len(_c.Stats)-1)

	LogRun("average_rxtx_KBps:" + fmt.Sprintf("%.2f", float64(rxBytes)/1000) + "<-->" + fmt.Sprintf("%.2f", float64(txBytes)/1000)) //Stats[]中的数据都是一秒一更新，所以单位为KBps
	LogRun("average_rxtxDrop_Package:" + fmt.Sprint(rxDropped) + "<-->" + fmt.Sprint(txDropped))

	// 所需数据在上边都已经计算完毕，以下为推送数据的代码
	if err := pushIt(fmt.Sprintf("%.2f", float64(rxBytes)/1000), timestamp, "net.rx.KBps", tags, containerID, "GAUGE", endpoint); err != nil {
		return err
	}
	if err := pushIt(fmt.Sprintf("%.2f", float64(txBytes)/1000), timestamp, "net.tx.KBps", tags, containerID, "GAUGE", endpoint); err != nil {
		return err
	}
	if err := pushIt(fmt.Sprint(rxDropped), timestamp, "net.rxDrop.packets", tags, containerID, "GAUGE", endpoint); err != nil {
		return err
	}
	if err := pushIt(fmt.Sprint(txDropped), timestamp, "net.txDrop.packets", tags, containerID, "GAUGE", endpoint); err != nil {
		return err
	}

	return nil

}
