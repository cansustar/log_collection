package common

// CollectEntry 要收集的日志的结构体
type CollectEntry struct {
	Path  string `json:"path"` // 指定了 JSON 序列化和反序列化时使用的字段名称。例如，JSON 中的 "path" 字段将与结构体中的 Path 字段相关联。
	Topic string `json:"topic"`
}
