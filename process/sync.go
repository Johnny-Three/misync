package process

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"wbproject/miusync/client"
	. "wbproject/miusync/logs"
	. "wbproject/miusync/structure"
)

var wg sync.WaitGroup

func DealReuqest(rein *Miu) ([]Miu, error) {

	reback := []Miu{}

	//先把时间字符串格式化成相同的时间类型
	t1, err1 := time.ParseInLocation("2006-01-02", rein.Fromdate, time.Local)
	if err1 != nil {
		return nil, err1
	}
	t2, err2 := time.ParseInLocation("2006-01-02", rein.Todate, time.Local)
	if err2 != nil {
		return nil, err2
	}

	//计算diff
	diff := int(t2.Sub(t1).Hours() / 24)

	//如果diff比7小，说明请求最近一周的数据，那么原样返回这个miu就可以了
	if diff <= 7 {

		reback = append(reback, *rein)
		return reback, nil
	}

	//如果diff大于31天，那么，fromdate从31天前开始计算
	if diff > 31 {

		diff = 31
		t1 = t2.Add(-1 * time.Hour * 24 * 31)
	}

	var rtmp Miu
	//从7天前开始计算
	for index := 0; index < diff/7; index++ {

		rtmp = Miu{rein.Userid, rein.LastuploadTime, rein.Appid, rein.Third_appid, rein.Third_appsecret, rein.Mac_key, rein.Call_id, rein.Access_token, rein.Fromdate, rein.Todate, rein.V, rein.L}
		rtmp.Fromdate = t1.Format("2006-01-02")
		rtmp.Todate = t1.Add(6 * 24 * time.Hour).Format("2006-01-02")
		t1 = t1.Add(7 * 24 * time.Hour)
		reback = append(reback, rtmp)
	}

	//考虑余数
	if !strings.EqualFold(rtmp.Todate, rein.Todate) {

		rein.Fromdate = t1.Format("2006-01-02")
		reback = append(reback, *rein)
	}

	return reback, nil
}

//需要同步的人群，开始并行同步，每次def并发量
func Sync(uids []*Miu, def int) {

	stepth := len(uids) / def
	//fmt.Println("stepth is: ", stepth)

	for i := 0; i < stepth; i++ {

		time.Sleep(1 * time.Millisecond)

		for j := i * def; j < (i+1)*def; j++ {

			wg.Add(1)

			go func(j int) {
				defer wg.Done()
				//todo .. 处理每个用户，从小米获取信息，处理信息并入库
				fmt.Println("hi,Sync is running in batch ")
				res, err := DealReuqest(uids[j])
				if err != nil {
					Logger.Critical(err)
					return
				}
				for i := 0; i < len(res); i++ {
					client.Post(&res[i])
				}

			}(j)
		}
		wg.Wait()
	}

	yu := len(uids) % def
	//模除部分处理
	if yu != 0 {

		for j := stepth * def; j < len(uids); j++ {

			time.Sleep(1 * time.Millisecond)

			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				//todo .. 处理每个用户，从小米获取信息，处理信息并入库
				fmt.Println("hi,Sync is running in yu ")
				res, err := DealReuqest(uids[j])
				if err != nil {
					fmt.Println(err)
					Logger.Critical(err)
					return
				}
				for i := 0; i < len(res); i++ {
					client.Post(&res[i])
				}
			}(j)
		}

		wg.Wait()

	}
}
