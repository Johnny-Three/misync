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
var version string = "1.0.0PR13"
var count int = 0

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

	Logger.Info("Misync running for once time ==================")

	//开启返回消息处理goroutine
	go HandleAnswer()

	//开启处理入库协程
	go func() {

		for {
			select {

			case m := <-User_walk_data_chan:
				err := InsertWalkDay(db, &m)
				if err != nil {
					fmt.Println("InsertWalkDay db", err)
					Logger.Critical("InsertWalkDay db", err)
				}
				//fmt.Println("insert walkdata")

				//todo..更新insertwalkhour..
				err = InsertWalkHour(db, &m)
				if err != nil {
					fmt.Println("InsertWalkHour db", err)
					Logger.Critical("InsertWalkHour db", err)
				}
				//fmt.Println("insert walkhour")

				//更新上次上传时间，取最近获取数据消息的那天，进行更新操作
				err = ModifyLastuploadtime(db, &m)
				//fmt.Println("modify lastupload time")
				if err != nil {
					fmt.Println("ModifyLastuploadtime db", err)
					Logger.Critical("ModifyLastuploadtime db", err)
				}
			}
		}

	}()

	//fmt.Println("interval is", interval)
	var timer *time.Ticker
	timer = time.NewTicker(time.Duration(1) * time.Millisecond)
	defer timer.Stop()

	//第一次跑..
	if count == 0 {

		count = 1
		start := time.Now()
		fmt.Println("Start to Load DB GetAllPerson... The Current time is ", start)
		Logger.Info("Start to Load DB GetAllPerson... The Current time is ", start)
		mius, err0 := GetAllPerson(db)
		fmt.Println("load db game over the len of mius is ", len(mius))
		Logger.Info("load db game over the len of mius is ", len(mius))
		elapsed := time.Since(start)
		fmt.Println("Load DB GetAllPerson query total time:", elapsed)
		Logger.Info("Load DB GetAllPerson query total time:", elapsed)
		if err0 != nil {
			Logger.Critical(err0.Error())
			return
		}

		Sync(mius, def)
	}

	if count == 1 {

		timer = time.NewTicker(time.Duration(interval) * time.Second)

		for {
			select {
			case <-timer.C:

				start := time.Now()
				fmt.Println("Start to Load DB GetAllPerson... The Current time is ", start)
				Logger.Info("Start to Load DB GetAllPerson... The Current time is ", start)
				mius, err0 := GetAllPerson(db)
				fmt.Println("load db game over the len of mius is ", len(mius))
				Logger.Info("load db game over the len of mius is ", len(mius))
				elapsed := time.Since(start)
				fmt.Println("Load DB GetAllPerson query total time:", elapsed)
				Logger.Info("Load DB GetAllPerson query total time:", elapsed)
				if err0 != nil {
					Logger.Critical(err0.Error())
					return
				}
				Sync(mius, def)
			}
		}
	}
}
