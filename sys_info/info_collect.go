package sysinfo

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	cli influxdb2.Client
	// a: 这是一个时间戳，用于记录上一次获取网络信息的时间
	lastNetIoStatTimeStamp int64
	// a: 这是一个结构体指针，用于记录上一次获取网络信息的数据
	lastNetInfo *NetInfo
)

func InitConnInflux() (client influxdb2.Client) {
	token := "Qa6oRmpcSNUKwDHuqt_KXFA1tEctpWiNeCPM8c28Bqic8IAnDkDUW1ccT_2Mqz8ISKVx7qYoNJfi0lQrXqXDSQ=="
	url := "http://localhost:8086"
	client = influxdb2.NewClient(url, token)
	return client
}

func WritesPoints(info DataInfo) {
	org := "cumt_life"
	bucket := "cumt_life_jcry_info"
	writeAPI := cli.WriteAPIBlocking(org, bucket)
	// 根据传入数据的类型，插入数据
	var fields map[string]interface{}
	var tags map[string]string
	var writeToDB = true // 标志，用于控制是否执行循环外的写入逻辑
	switch info.Data.(type) {
	case CpuInfo:
		// 这行代码的意思是：将 info.Data 转换为 CpuInfo 这是 Go 语言中的类型断言写法。使用 .(Type) 的形式，其中 Type 是你希望将数据转换成的类型。
		data := info.Data.(CpuInfo) // 因为Data字段是interface{}， 在使用该数据时需要进行类型断言，将其转换成具体的类型。
		fields = map[string]interface{}{
			"cpu_percent": data.CpuPercent,
		}
		tags = nil
	case MemInfo:
		data := info.Data.(MemInfo)
		fields = map[string]interface{}{
			"total":        data.Total,
			"available":    data.Available,
			"used":         data.Used,
			"used_percent": data.UsedPercent,
			"free":         data.Free,
			"buffers":      data.Buffers,
			"cached":       data.Cached,
		}
		tags = nil
	case DiskInfo:
		data := info.Data.(DiskInfo)
		// 这里的k是分区的名字，v是分区的信息
		for k, v := range data.PartitionUsageStat {
			fields = map[string]interface{}{
				"total":               v.Total,
				"free":                v.Free,
				"used":                v.Used,
				"used_percent":        v.UsedPercent,
				"inodes_total":        v.InodesTotal,
				"inodes_used":         v.InodesUsed,
				"inodes_free":         v.InodesFree,
				"inodes_used_percent": v.InodesUsedPercent,
			}
			// tags中的数据是用来过滤数据的，比如我们想要查询某一个分区的数据，那么就可以使用tags中的数据进行过滤
			tags = map[string]string{
				"partition": k,
			}
			point := write.NewPoint(info.Measurement, tags, fields, time.Now())
			time.Sleep(1 * time.Second) // separate points by 1 second
			if err := writeAPI.WritePoint(context.Background(), point); err != nil {
				log.Fatal(err)
			}
			fmt.Println(info.Measurement, "insert success")
		}
		writeToDB = false
	case NetInfo:
		data := info.Data.(NetInfo)
		for k, v := range data.NetIOCounterStat {
			fields = map[string]interface{}{
				"bytes_sent_rate":   v.BytesSentRate,
				"bytes_recv_rate":   v.BytesRecvRate,
				"packets_sent_rate": v.PacketsSentRate,
				"packets_recv_rate": v.PacketsRecvRate,
			}
			tags = map[string]string{
				"name": k,
			}
			point := write.NewPoint(info.Measurement, tags, fields, time.Now())
			time.Sleep(1 * time.Second) // separate points by 1 second
			if err := writeAPI.WritePoint(context.Background(), point); err != nil {
				log.Fatal(err)
			}
			fmt.Println(info.Measurement, "insert success")
		}
		writeToDB = false

	}
	if writeToDB {
		point := write.NewPoint(info.Measurement, tags, fields, time.Now())

		time.Sleep(1 * time.Second) // separate points by 1 second
		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Fatal(err)
		}
		fmt.Println(info.Measurement, "insert success")
	}
}

func getCpuInfo() {
	percent, _ := cpu.Percent(time.Second, false)
	var info DataInfo
	info.Measurement = "device_cpu_info"
	var cpuInfo CpuInfo
	cpuInfo.CpuPercent = percent[0]
	println(cpuInfo.CpuPercent)
	info.Data = cpuInfo
	WritesPoints(info)
}

func getMemInfo() {
	var info DataInfo
	info.Measurement = "device_mem_info"

	var memInfo MemInfo
	data, err := mem.VirtualMemory()

	if err != nil {
		fmt.Printf("get mem info failed, err:%v", err)
		return
	}

	memInfo.Total = data.Total
	memInfo.Available = data.Available
	memInfo.Used = data.Used
	memInfo.UsedPercent = data.UsedPercent
	memInfo.Free = data.Free
	memInfo.Buffers = data.Buffers
	memInfo.Cached = data.Cached
	info.Data = memInfo
	WritesPoints(info)
}

func getDiskInfo() {
	var info DataInfo
	info.Measurement = "device_disk_info"
	var diskInfo DiskInfo
	// make的第二个参数表示的是map的容量，而不是长度
	diskInfo.PartitionUsageStat = make(map[string]*disk.UsageStat, 16)
	// partitions中的是每一个分区的信息
	partitions, err := disk.Partitions(true)
	if err != nil {
		fmt.Printf("get partitions failed, err:%v", err)
		return
	}
	// 遍历每一个分区，获取每一个分区的信息
	for _, partition := range partitions {
		// usageStat中是每一个分区的信息,参数partition.Mountpoint是分区的挂载点
		// 挂载点是指将外部设备（如U盘、移动硬盘等）与计算机的文件系统进行连接的一个目录
		usageStat, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			fmt.Printf("get partition %s usage failed, err:%v", partition.Mountpoint, err)
			continue
		}
		diskInfo.PartitionUsageStat[usageStat.Path] = usageStat
	}
	info.Data = diskInfo
	WritesPoints(info)
}

func getNetInfo() {
	var info DataInfo
	info.Measurement = "device_net_info"
	netIOs, err := net.IOCounters(true)
	if err != nil {
		fmt.Printf("get net io counters failed, err:%v", err)
		return
	}
	var netInfo NetInfo
	netInfo.NetIOCounterStat = make(map[string]*IoStat, 16)
	// 获取当前时间
	nowTimeStamp := time.Now().Unix()

	for _, netIO := range netIOs { // 每个网卡
		var ioStat = new(IoStat)
		ioStat.BytesSent = netIO.BytesSent
		ioStat.BytesRecv = netIO.BytesRecv
		ioStat.PacketsSent = netIO.PacketsSent
		ioStat.PacketsRecv = netIO.PacketsRecv
		// 将具体网卡数据的ioStat变量添加到netInfo中
		// q: 为什么这里不能放到continue 后面？
		//a: 因为如果放到continue后面，那么就会导致netInfo中的数据不完整，因为如果当前网卡的数据获取失败，那么就会跳过这个网卡，这样就会导致netInfo中的数据不完整
		netInfo.NetIOCounterStat[netIO.Name] = ioStat // 不要放到continue后面
		// 计算速率
		if lastNetIoStatTimeStamp == 0 || lastNetInfo == nil {
			continue
		}
		// 获取时间间隔
		interval := nowTimeStamp - lastNetIoStatTimeStamp
		BytesSentRate := float64(ioStat.BytesSent-lastNetInfo.NetIOCounterStat[netIO.Name].BytesSent) / float64(interval)
		BytesRecvRate := float64(ioStat.BytesRecv-lastNetInfo.NetIOCounterStat[netIO.Name].BytesRecv) / float64(interval)
		PacketsSentRate := float64(ioStat.PacketsSent-lastNetInfo.NetIOCounterStat[netIO.Name].PacketsSent) / float64(interval)
		PacketsRecvRate := float64(ioStat.PacketsRecv-lastNetInfo.NetIOCounterStat[netIO.Name].PacketsRecv) / float64(interval)
		ioStat.BytesSentRate = BytesSentRate
		ioStat.BytesRecvRate = BytesRecvRate
		ioStat.PacketsSentRate = PacketsSentRate
		ioStat.PacketsRecvRate = PacketsRecvRate

		// q: 为什么这里要用指针？
		// a: 因为netIO是一个结构体，如果不用指针的话，那么在循环中每次都会创建一个新的结构体，这样就会导致每次循环中的结构体都是新的，而不是原来的那个结构体
	}
	// 将当前的时间保存为上一次的时间
	lastNetIoStatTimeStamp = nowTimeStamp
	// 将当前的数据保存为上一次的数据
	lastNetInfo = &netInfo

	info.Data = netInfo
	WritesPoints(info)

}
func CollectSysInfo() {
	cli = InitConnInflux()

	// 启动 getCpuInfo 和 getMemInfo 的 goroutine
	go func() {
		ticker := time.Tick(time.Second)
		for {
			select {
			case <-ticker:
				getNetInfo()
			}
		}
	}()
	go func() {
		ticker := time.Tick(time.Second)
		for {
			select {
			case <-ticker:
				getDiskInfo()
			}
		}
	}()
	go func() {
		ticker := time.Tick(time.Second)
		for {
			select {
			case <-ticker:
				// q: 你猜我为啥要把他俩放一个goroutine中
				// q: 如果把他俩和getNetInfo放在一个goroutine中，getnet和getdisk中有一个执行时间很长，那么就会导致getCpuInfo和getMemInfo的执行时间也很长，因为他俩是串行的
				getCpuInfo()
				getMemInfo()
			}
		}
	}()
	select {}
}
