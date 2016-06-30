package process

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"strconv"
	"strings"
	"time"
	"wbproject/miusync/client"
	. "wbproject/miusync/dbhelper"
	. "wbproject/miusync/envbuild"
	. "wbproject/miusync/logs"
	. "wbproject/miusync/structure"
	"wbproject/miusync/util"
)

var User_walk_data_chan chan User_walkdays_struct

func init() {

	User_walk_data_chan = make(chan User_walkdays_struct, 1024)
	Map = util.NewBeeMap()
	MapOld = util.NewBeeMap()
}

func GetTimestamp(date string) (timestamp int64) {
	tm, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	timestamp = tm.Unix()
	return timestamp
}

func Decode(msg Reback) error {

	js, err := simplejson.NewJson([]byte(msg.JsonCode))
	if err != nil {
		errback := fmt.Sprintf("decode json error the error msg is %s", err.Error())
		return errors.New(errback)
	}

	var ad AnswerData
	walkdays := []AnswerData{}
	//Logger.Debug("in decode the msg is", msg)
	//Logger.Info("in decode the msg is", msg)
	arr, _ := js.Get("data").Array()
	for index, _ := range arr {

		//walkdate string到timestamp转化
		walkdate := js.Get("data").GetIndex(index).Get("date").MustString()
		ad.Walkdate = GetTimestamp(walkdate)

		var tmp int64 = time.Now().Unix()
		var currentdate string = time.Unix(tmp, 0).Format("2006-01-02")
		//判断当前时间是否和传回的时间相等，如果相等，则从内存中拿到对应的step数据，如果step没有变化，
		//则后续操作全免除，如果内存中没有用户对应的数据，则在数据入库后需要将数据更新过去；
		//如果有变化，入库后数据需要更新过去；
		//fmt.Println(currentdate, walkdate)
		if strings.EqualFold(currentdate, walkdate) {

			//step
			step := js.Get("data").GetIndex(index).Get("step").MustString()
			st, est := strconv.Atoi(step)
			if est != nil {
				fmt.Println("step 字符串转换成整数失败")
			}

			// Retrieve steps from map.
			if tmp, ok := Map.Get(msg.Userid); ok {
				//Get Steps,if no change then won't do anyting ..
				if st == tmp.(int) {

					fmt.Printf("用户：%d 日期：%s，步数值:%s 与上次获取时无变化，不更新表\n", msg.Userid, walkdate, step)
					return nil
					//插入成功后再更新这个新值么？异常回退流程值得商榷
				} else {

					Logger.Info("Get msg:", msg)
					MapOld.Set(msg.Userid, tmp)
					Map.Set(msg.Userid, st)
				}
			} else {

				Logger.Info("Get msg:", msg)
				//混沌初开的时候并不存在map值,mapold里面放入0
				Map.Set(msg.Userid, st)
				MapOld.Set(msg.Userid, 0)
			}
		}

		//walkdistance 米到厘米转化*1000
		walkdistance := js.Get("data").GetIndex(index).Get("walkDistance").MustString()
		wd, edt := strconv.Atoi(walkdistance)
		if edt != nil {
			fmt.Println("walkdistance 字符串转换成整数失败")
		}
		ad.Walkdistance = wd * 1000

		//walktime
		walktime := js.Get("data").GetIndex(index).Get("walkTime").MustString()
		wt, ekt := strconv.Atoi(walktime)
		if ekt != nil {
			fmt.Println("walktime 字符串转换成整数失败")
		}
		runtime := js.Get("data").GetIndex(index).Get("runTime").MustString()
		rt, ert := strconv.Atoi(runtime)
		if ert != nil {
			fmt.Println("runtime 字符串转换成整数失败", err)
		}
		ad.Walktime = wt + rt

		//calorieconsumed  千卡转换*1000
		calorie, ect := strconv.ParseFloat(js.Get("data").GetIndex(index).Get("calorie").MustString(), 64)
		if ect != nil {
			fmt.Println("calorie 字符串转换成float失败")
		}
		c, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", calorie), 2)
		ad.Calorieconsumed = c

		//fatconsumed 千卡转换*1000
		k := calorie / 7
		f, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", k), 2)
		ad.Fatconsumed = f

		//step
		step := js.Get("data").GetIndex(index).Get("step").MustString()
		st, est := strconv.Atoi(step)
		if est != nil {
			fmt.Println("step 字符串转换成整数失败")
		}
		ad.Stepnumber = st

		ex, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(st)*float64(0.000556)), 2)
		ad.Exerciseamount = ex
		//填充ad的stepwidth,weight,stepgoal信息..
		_ = SetAnswerData(GetDB(), &ad, msg.Userid)
		//用户此次Post返回的数据消息存储起来..
		walkdays = append(walkdays, ad)
	}

	//walkdays长度为0
	if len(walkdays) == 0 {
		return nil
	}

	var uc User_walkdays_struct
	uc.Uid = msg.Userid
	uc.LastuploadTime = msg.LastuploadTime
	uc.Walkdays = walkdays
	//fmt.Println("uc is", uc)

	User_walk_data_chan <- uc

	return nil
}

func HandleAnswer() error {

	for {

		msg := <-client.Post_request_chan

		js, err := simplejson.NewJson([]byte(msg.JsonCode))
		if err != nil {
			errback := fmt.Sprintf("decode json error the error msg is %s", err.Error())
			return errors.New(errback)
		}

		code := js.Get("code").MustInt()

		switch {

		case code == 0:
			//这7天没有数据

		case code == 1:
			err := Decode(msg)
			if err != nil {
				Logger.Critical(err)
			}

		case code == -40000:
			fmt.Println("code is -40000!")
			ModifyStatus(GetDB(), msg.Userid)

		default:
			//其余情况，需要记录异常返回消息
			Logger.Critical(msg)
		}
	}

	return nil
}
