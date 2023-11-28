package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"log_collection/kafka"
	"log_collection/tailfile"
	"strings"
	"time"
)

// 日志收集的客户端
// 类似的开源项目还有filebeat
// 收集指定目录下的日志文件，并发送到kafka中

type Config struct {
	KafkaConfig   `ini:"kafka"` // ini标签的值要和config.ini中的节相同
	CollectConfig `ini:"collect"`
}

type KafkaConfig struct {
	Address  string `ini:"address"`
	Topic    string `ini:"topic"`
	ChanSize int64  `ini:"chan_size"`
}

type CollectConfig struct {
	LogFilePath string `ini:"logfile_path"`
}

// 真正的业务逻辑
func run() (err error) {
	// TailObj --> log --> Client --> kafka

	for true {
		line, ok := <-tailfile.TailObj.Lines
		if !ok {
			logrus.Warning("tail file close reopen, filename:%s\n", tailfile.TailObj.Filename)
			time.Sleep(time.Second)
			continue
		}
		// 如果是空行，略过
		fmt.Println(line.Text)
		if len(strings.Trim(line.Text, "\r")) == 0 {
			logrus.Info("出现空行，跳过..")
			continue
		}
		// 将日志发送到kafka
		// 利用通道，将同步的代码改为异步的
		// 把读出来的一行日志，包装成kafka里的msg类型，丢到通道中
		msg := &sarama.ProducerMessage{}
		msg.Topic = "web_log"
		msg.Value = sarama.StringEncoder(line.Text)
		// 丢到管道中:为什么要丢到通道里，而不是直接调用client.SendMessage(msg)
		// 因为如果直接调用SendMessage的话，相当于for循环里，取一行日志，然后往kafka中发送一次。当数据量比较大的时候，for循环压力比较大
		// 通过一个通道，把日志包装成msg。设计Channel时，不是直接使用String,而是使用内存地址。占用空间比较小，可以开更多的channel.
		// 这里暴露了整个Channel, 但是只希望将消息发送到通道中，而消息的读取是消费者读kafka实现的，而不是读Channel，所以MsgChan最好是私有的.可以写一个函数返回msgChan
		//kafka.MsgChan <- msg
		kafka.ToMsgChan(msg) // 这里的ToMsgChan被封装成了函数
	}
	return
}

func main() {
	// 0. 读配置文件 go-ini， 加载kafka和collect的配置项
	// 下面要用反射给结构体赋值，要改变结构体变量的值，go中函数传参都是拷贝，用new，可以得到一个结构体指针
	var configObj = new(Config)
	err := ini.MapTo(configObj, "./conf/config.ini")
	if err != nil {
		logrus.Error("load config failed, err: %v", err)
		return
	}
	// "%#v" 这个格式字符串，它用于打印 Go 语言的值的 Go 语法表示形式。具体而言，%#v 会以 Go 语法格式输出变量的值，包括类型信息和字段名（如果是结构体）。
	fmt.Printf("%#v\n", configObj)
	// 1. 初始化（做好准备工作） 连接kafka，检查tail连接是否正常  连接kafka,初始化msgChan,启后台goroutine往kafka发msg
	err = kafka.Init([]string{configObj.KafkaConfig.Address}, configObj.KafkaConfig.ChanSize)
	if err != nil {
		logrus.Error("init kafka failed, err:%v", err)
		return
	}
	logrus.Info("init kafka success!")
	// 2. 根据配置中的日志路径，创建对应的tailobj，使用tail去收集日志
	// 从结构体中加载对象
	err = tailfile.Init(configObj.CollectConfig.LogFilePath)
	if err != nil {
		logrus.Error("init tailfile failed, err:%v", err)
		return
	}
	logrus.Info("init tailfile success!")
	// 3. 把日志通过sarama发往kafka
	// 从tailobj.lines中一行行读日志，包装成kafka msg,丢到MsgChan中
	err = run()
	if err != nil {
		logrus.Error("run failed, err: %v", err)
		return
	}
}
