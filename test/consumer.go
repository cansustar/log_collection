package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"sync"
)

func consumer_test() {
	consumer, err := sarama.NewConsumer([]string{"127.0.0.1:9092"}, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		return
	}
	partitionList, err := consumer.Partitions("web_log") // 根据topic取到所有的分区
	if err != nil {
		fmt.Printf("fail to get list of partition: err:%v\n", err)
		return
	}
	fmt.Println(partitionList)
	var wg sync.WaitGroup

	for partition := range partitionList {
		// 针对每个分区创建一个对应的分区消费者，读取时更快
		pc, err := consumer.ConsumePartition("web_log", int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("failed to start consumer for partition %d, err: %v\n", partition, err)
		}
		defer pc.AsyncClose()
		wg.Add(1)
		// 异步从每个分区消费信息
		go func(sarama.PartitionConsumer) {
			// go创建的goroutine的函数相当于闭包，要访问外面的变量，需要捕获变量，捕获的变量就在go func {} ()的（）里
			for msg := range pc.Messages() {
				fmt.Printf("partition: %d Offset: %d key: %s Value: %s\n", msg.Partition, msg.Offset, msg.Key, msg.Value)
			}
		}(pc)
	}
	wg.Wait()
}
