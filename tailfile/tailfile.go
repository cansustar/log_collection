package tailfile

import (
	"github.com/nxadm/tail"
	"github.com/sirupsen/logrus"
)

// tail相关

var (
	TailObj *tail.Tail
)

func Init(filename string) (err error) {
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	// 打开文件，读取数据
	TailObj, err = tail.TailFile(filename, config) // tails要作为文件供外部使用，所以要在该包内var声明
	if err != nil {
		/*
			格式化字符串输出时：
			%s 用于字符串，而 %v 用于将值以默认的方式打印。两者的主要区别在于 %v 是一个通用的格式化标识符，会根据值的类型选择适当的格式化方式，而 %s 专门用于字符串。
			在你的例子中，%s 用于打印字符串 filename，而 %v 用于打印变量 err。这样的选择主要取决于你想如何显示这两个值。
			如果你知道 err 是一个字符串，你可以使用 %s，但如果 err 是一个任意类型的错误值，你可能更愿意使用 %v，以便 logrus 根据 err 的实际类型选择适当的显示方式。
		*/
		logrus.Error("tailfile: create tailobj for path %s faild, err:%v\n", filename, err)
		return
	}
	return
}
