package common

// CollectEntry 要收集的日志的结构体
type CollectEntry struct {
	// 去哪个路径读取文件
	Path string `json:"path"` // 指定了 JSON 序列化和反序列化时使用的字段名称。例如，JSON 中的 "path" 字段将与结构体中的 Path 字段相关联。
	// 日志文件发往kafka的Topic
	Topic string `json:"topic"`
}
