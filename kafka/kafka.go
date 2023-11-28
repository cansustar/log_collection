package kafka

import (
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

var (
	// 小写： 只有当前包内可以用 大写： 外部包也可以用
	client  sarama.SyncProducer          // 这里client类型的确定: 连接kafka时sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, config)， NewSyncProducer最终的返回类型是SyncProducer
	msgChan chan *sarama.ProducerMessage // msgChan是一个通道，关于其类型，因为在msg发送给kafka时，封装消息所使用的msg := &sarama.ProducerMessage{}  实际上是一个指针
)

// Init kafka相关操作
// Init 用来初始化全局的kafka Client
func Init(address []string, chanSize int64) (err error) {

	// go中，如果指定了返回值的话，需要把返回值和返回类型用括号括起来
	//// 1. 生产者配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // ack
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 随机的一个分区
	config.Producer.Return.Successes = true                   //确认
	// 3. 连接kafka
	client, err = sarama.NewSyncProducer(address, config) // 因为要给别的包调用client,所以在上面用var声明一个client
	if err != nil {
		logrus.Error("kafka: producer closed, err:", err)
		return
	}
	// 初始化msgChan
	msgChan = make(chan *sarama.ProducerMessage, chanSize)
	// 初始化时，起一个后台的goroutine，从msgChan中读取数据,发送到kafka
	go sendMsg()
	return
}

// 从通道 msgChan 中 读取消息，发送给kafka
func sendMsg() {
	for true {
		select {
		case msg := <-msgChan:
			// 发送给kafka
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				logrus.Warning("send msg failed, err:", err)
				return
			}
			logrus.Infof("send msg to kafka success. pid:%v  offset:%v", pid, offset)
		}
	}
}

// ToMsgChan MsgChan 定义一个函数向外暴露msgChan, chan<- 单向通道
func ToMsgChan(msg *sarama.ProducerMessage) {
	msgChan <- msg
}
