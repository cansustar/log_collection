# log_collection
使用tailfile收集日志，采集系统运行指标。发送到kafka或存储到influxdb中。

# 整体设计
1. 使用go-ini 读取etcd,kafka等的配置信息并映射成结构体，用于初始化influxdb,kafka,etcd的连接。
2. 使用etcd来管理日志文件配置，各个服务器上部署该服务后，根据本机ip拼接到collect_log_%s_conf作为key，value为json格式的待采集的日志topic与路径（topic服务于kafka）
3. 根据从etcd中读取到的配置信息，启动相应的tailfileTask，监听日志文件，每当有新的日志信息时发送到kafka.
4. 使用watch来监听etcd内相应key的变化。当value变化时（添加或移除），通过channel告知tailMange，清除或开启相应的tailfileTask。
5. 系统指标采集后直接存储到influxdb中，并用其可视化功能展示。


## 启动前需要准备好的一些中间件
### kafka
使用kafka作为MQ来传输日志数据
### etcd
使用etcd来作为配置中心，各服务器从etcd上读取自己服务器所需要采集日志的路径配置。
### influxdb
作为时序数据库，存储并用其可视化功能展示系统指标。
