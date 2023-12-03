package sysinfo

import "github.com/shirou/gopsutil/disk"

// DataInfo 总的结构体，包含了数据的类型和数据  主要用于确定数据的尺度（表）
type DataInfo struct {
	Measurement string
	Data        interface{} // 类型是interface{} 表示可以存储任意类型的数据
}

type CpuInfo struct {
	CpuPercent float64 `json:"cpu_percent"`
}

// MemInfo 内存信息
type MemInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
	Free        uint64  `json:"free"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
}

// 磁盘信息
type DiskInfo struct {
	PartitionUsageStat map[string]*disk.UsageStat `json:"partition_usage_stat"`
}

// NetInfo 网络信息
type NetInfo struct {
	NetIOCounterStat map[string]*IoStat `json:"net_iocouter_stat"`
}

type IoStat struct {
	// 计算速率的时候需要用到上一次的数据，所以需要记录上一次的数据
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64

	BytesSentRate   float64 `json:"bytes_sent_rate"`
	BytesRecvRate   float64 `json:"bytes_recv_rate"`
	PacketsSentRate float64 `json:"packets_sent_rate"`
	PacketsRecvRate float64 `json:"packets_recv_rate"`
}
