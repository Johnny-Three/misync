package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	. "wbproject/miusync/dbhelper"
	. "wbproject/miusync/envbuild"
	. "wbproject/miusync/logs"
	. "wbproject/miusync/process"
)

var err error
var version string = "1.0.0PR5"

func main() {

	args := os.Args

	if len(args) == 2 && (args[1] == "-v") {

		fmt.Println("看好了兄弟，现在的版本是【", version, "】，可别弄错了")
		os.Exit(0)
	}

	flag.Parse()

	db, interval, def, err := EnvBuild()
	if err != nil {
		panic(err.Error())
	}

	//开启返回消息处理goroutine
	go HandleAnswer()

	//开启处理入库协程
	go func() {

		for {
			select {

			case m := <-User_walk_data_chan:
				err := InsertWalkDay(db, &m)
				if err != nil {
					fmt.Println("insert db", err)
				}
				//todo..更新insertwalkhour..
				err = InsertWalkHour(db, &m)
				if err != nil {
					fmt.Println("insert db", err)
				}
				//更新上次上传时间，取最近获取数据消息的那天，进行更新操作
				err = ModifyLastuploadtime(db, &m)
				if err != nil {
					fmt.Println("insert db", err)
				}
			}
		}

	}()

	fmt.Println("interval is", interval)
	timer := time.NewTicker(time.Duration(1) * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			err = ModifyPerson(db)
			if err != nil {
				Logger.Critical(err.Error())
				return
			}
			mius, err0 := GetAllPerson(db)
			if err0 != nil {
				Logger.Critical(err0.Error())
				return
			}
			//fmt.Println("mius is", mius[0])
			Sync(mius, def)
			timer = time.NewTicker(time.Duration(interval) * time.Second)

		default:
			time.Sleep(1 * time.Second)
			//fmt.Println("休息2s，继续工作！")
		}
	}
}
