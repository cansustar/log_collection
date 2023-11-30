package tailfile

import (
	"github.com/sirupsen/logrus"
	"log_collection/common"
)

// tailTask 的管理者

type tailTaskMgr struct {
	tailTaskMap      map[string]*tailTask       // 所有的tailTask任务  map[key]value  将所有收集日志的小弟管理到这个map中
	collectEntryList []common.CollectEntry      // 所有配置项
	confChan         chan []common.CollectEntry // 等待新配置的通道
}

var (
	ttMgr *tailTaskMgr
)

func (t *tailTaskMgr) watch() { // tailTaskMgr 类型的一个实例方法，可以通过 ttMgr.watch() 这样的方式调用。
	for true {
		// 派一个小弟等着新配置，
		newConf := <-t.confChan // 说明新的配置来了
		// 新配置来了之后，应该管理一下之前启动的哪些tailTask
		logrus.Infof("get new conf from etcd, conf:%v, start manager", newConf)
		for _, conf := range newConf {
			// 原来已经存在的tailTask任务，不用动
			if t.isExist(conf) {
				continue
			}
			// 原来没有的，要新创建一个tailTask任务
			tt := newTailTask(conf.Path, conf.Topic)
			err := tt.Init()
			if err != nil {
				logrus.Errorf("create tailobj for path:%s failed, err:%v", conf.Path, err)
				continue
			}
			t.tailTaskMap[tt.path] = tt
			// 起一个后台的goroutine收集日志
			go tt.run()
		}
		// 原来有，现在没有的，要删除tailTask任务
		// 找出tailTaskMap中存在，但是newConf中不存在的那些tailTask， 把他们都停掉
		for key, task := range t.tailTaskMap {
			var found bool
			for _, conf := range newConf {
				if key == conf.Path {
					found = true
					break
				}
			}
			if !found {
				//  这个task要被停掉
				logrus.Infof("the task collect path:%s need to stop", task.path)
				delete(t.tailTaskMap, key) //将该配置从管理类中删掉
				task.cancel()
			}
		}
	}
}
func (t *tailTaskMgr) isExist(conf common.CollectEntry) bool {
	_, ok := t.tailTaskMap[conf.Path]
	return ok
}

func Init(allConf []common.CollectEntry) (err error) {
	// allConf中存了若干个日志的收集项，
	// 针对每一个日志收集项创建一个对应的tailObj
	ttMgr = &tailTaskMgr{
		tailTaskMap:      make(map[string]*tailTask, 50),
		collectEntryList: allConf,
		confChan:         make(chan []common.CollectEntry),
	}
	for _, conf := range allConf {
		tt := newTailTask(conf.Path, conf.Topic) // 创建一个日志收集任务
		err = tt.Init()
		if err != nil {
			logrus.Errorf("create tailobj for path:%s failed, err:%v", conf.Path, err)
			continue
		}
		logrus.Info("create a tail task for path:%s success", conf.Path)
		ttMgr.tailTaskMap[tt.path] = tt // 把创建的这个tail task 登记到ttmgr中
		go tt.run()                     // 调用上面定义的tailTask的run方法，有几个配置项启动几个goroutine
	}
	// 在后台等新的配置
	go ttMgr.watch()
	return
}

func SendNewConf(newConf []common.CollectEntry) {
	ttMgr.confChan <- newConf

}
