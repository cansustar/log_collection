package common

import (
	"net"
	"strings"
)

// CollectEntry 要收集的日志的结构体
type CollectEntry struct {
	// 去哪个路径读取文件
	Path string `json:"path"` // 指定了 JSON 序列化和反序列化时使用的字段名称。例如，JSON 中的 "path" 字段将与结构体中的 Path 字段相关联。
	// 日志文件发往kafka的Topic
	Topic string `json:"topic"`
}

// GetOutboundIP 获取本机ip的函数
func GetOutboundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
		// log.Fatal(err) // 这语句会打印错误信息到标准错误输出，然后调用 os.Exit(1) 终止程序。
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)      // 10.11.0.210:54514
	ip = strings.Split(localAddr.IP.String(), ":")[0] // 去掉端口 但实际上localAddr才有端口，才需要这样分割
	return
}
