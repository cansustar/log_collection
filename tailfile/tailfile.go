package tailfile

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/nxadm/tail"
	"github.com/sirupsen/logrus"
	"log_collection/common"
	"log_collection/kafka"
	"time"
)

// tail相关

type tailTask struct { // 有一个配置项，就要创建这样的一个结构体的实例
	path   string
	topic  string
	tObj   *tail.Tail
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	confChan chan []common.CollectEntry
)

//var (
//	TailObj *tail.Tail
//)

func newTailTask(path, topic string) *tailTask {
	ctx, cancel := context.WithCancel(context.Background())
	tt := &tailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cancel,
	}

	return tt
}

func (t *tailTask) Init() (err error) { // 这是一个属于tailTask类型的Init方法，不接收参数，返回一个错误
	cfg := tail.Config{ // cfg是公用的tail配置，可以在循环外
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	t.tObj, err = tail.TailFile(t.path, cfg)
	return
}

func (t *tailTask) run() {
	// 读取日志，发往kafka
	logrus.Info("collect for path: %s is running...", t.path)
	for true {
		select {
		case <-t.ctx.Done(): // 只要调用t.cancel() 就会收到信号
			logrus.Infof("path: %s is stopping", t.path)
			t.tObj.Cleanup() // 当删除一个日志配置后，对于不需要再监听的日志，需要清理tail对象
			return
		case line, ok := <-t.tObj.Lines:
			if !ok {
				// 待优化，这里打开文件失败会一直重启
				logrus.Warningf("tail file close reopen, filename:%s", t.path)
				time.Sleep(time.Second)
				continue
			}
			// 将日志发送到kafka
			// 利用通道，将同步的代码改为异步的
			// 把读出来的一行日志，包装成kafka里的msg类型，丢到通道中
			msg := &sarama.ProducerMessage{}
			msg.Topic = t.topic
			msg.Value = sarama.StringEncoder(line.Text)
			// 丢到管道中:为什么要丢到通道里，而不是直接调用client.SendMessage(msg)
			// 因为如果直接调用SendMessage的话，相当于for循环里，取一行日志，然后往kafka中发送一次。当数据量比较大的时候，for循环压力比较大
			// 通过一个通道，把日志包装成msg。设计Channel时，不是直接使用String,而是使用内存地址。占用空间比较小，可以开更多的channel.
			// 这里暴露了整个Channel, 但是只希望将消息发送到通道中，而消息的读取是消费者读kafka实现的，而不是读Channel，所以MsgChan最好是私有的.可以写一个函数返回msgChan
			//kafka.MsgChan <- msg
			kafka.ToMsgChan(msg) // 这里的ToMsgChan被封装成了函数

		}
	}
}
