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
var version string = "1.0.0PR11"

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

	fmt.Println("interval is", interval)
	timer := time.NewTicker(time.Duration(1) * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:

			start := time.Now()
			mius, err0 := GetAllPerson(db)
			fmt.Println("load db game over the len of mius is", len(mius))
			Logger.Info("load db game over the len of mius is", len(mius))
			elapsed := time.Since(start)
			fmt.Println("Load db person query total time:", elapsed)
			Logger.Info("load db game over the len of mius is", len(mius))
			if err0 != nil {
				Logger.Critical(err0.Error())
				return
			}
			Sync(mius, def)
			timer = time.NewTicker(time.Duration(interval) * time.Second)

		default:
			time.Sleep(1 * time.Second)
			//fmt.Println("休息2s，继续工作！")
		}
	}
}
