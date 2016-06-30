package dbhelper

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"time"
	//. "wbproject/miusync/process"
	. "wbproject/miusync/structure"
	"wbproject/miusync/util"
)

var miuserial = "068000001110000000000000"
var hourstep, moyu, h23, hc int

func InsertWalkHour(db *sql.DB, uc *User_walkdays_struct) error {

	enddate := uc.Walkdays[len(uc.Walkdays)-1].Walkdate
	begindate := uc.LastuploadTime

	t, _ := time.ParseInLocation("20060102", time.Now().Format("20060102"), time.Local)
	currentdate := t.Unix()
	//分三种情况处理 1.跨天  2.天内跨小时 3.小时内
	//跨天所有数据写h23，天内跨小时写当前小时（直接写）、小时内同样写当前小时(增量写)

	var spanday bool

	//跨天的第一个条件：修改了lastuploadtime为昨天且今天无数据，相当于重新写昨天的小时数据，有很大的风险，因为这将会重写昨天的小时数据到23点)  (currentdate != enddate && enddate == begindate) ，考虑后还是先不加，本身逻辑有问题且不合乎事宜
	//问题描述：今日无数据，昨日有数据，修改lastuploadtime为昨天，因为spanday不跨天,会进入小时增量的分支

	//跨天
	if util.DaysDiff(enddate, begindate) > 0 {

		spanday = true

		sqlStr := `
	   INSERT INTO wanbu_data_walkhour (userid, walkdate, timestamp,servertime, hour0, hour1, hour2, hour3, hour4, hour5, hour6, hour7, hour8, hour9, hour10, hour11, hour12, hour13, hour14, hour15, hour16, hour17, hour18, hour19, hour20, hour21, hour22, hour23, hour24, hour25 ) VALUES `

		vals := []interface{}{}

		for _, uds := range uc.Walkdays {

			var hour HourData
			hour.Init()
			//当前天写当前小时
			if currentdate == uds.Walkdate {
				AssginHourData(&hour, &uds)
			} else {
				//当前天之前写23点
				hour.H23 = uds.GetHourString()
			}

			sqlStr += "(?,?,UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),"

			vals = append(vals, uc.Uid, uds.Walkdate, hour.H0, hour.H1, hour.H2, hour.H3, hour.H4, hour.H5, hour.H6, hour.H7, hour.H8, hour.H9, hour.H10, hour.H11, hour.H12, hour.H13, hour.H14, hour.H15, hour.H16, hour.H17, hour.H18, hour.H19, hour.H20, hour.H21, hour.H22, hour.H23, "0,0,0,0,0,0", "0,0,0,0,0,0")
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]

		sqlStr += `ON DUPLICATE KEY UPDATE timestamp = VALUES(timestamp),servertime = VALUES(servertime),hour0 = VALUES(hour0),hour1 = VALUES(hour1),hour2 = VALUES(hour2),hour3 = VALUES(hour3),hour4 = VALUES(hour4),hour5 = VALUES(hour5),hour6 = VALUES(hour6),hour7 = VALUES(hour7),hour8 = VALUES(hour8),hour9 = VALUES(hour9),hour10 = VALUES(hour10),hour11 = VALUES(hour11),hour12 = VALUES(hour12),hour13 = VALUES(hour13),hour14 = VALUES(hour14),hour15 = VALUES(hour15),hour16 = VALUES(hour16),hour17 = VALUES(hour17),hour18 = VALUES(hour18),hour19 = VALUES(hour19),hour20 = VALUES(hour20),hour21 = VALUES(hour21),hour22 = VALUES(hour22),hour23 = VALUES(hour23)`

		//format all vals at once
		_, err := db.Exec(sqlStr, vals...)

		if err != nil {
			fmt.Println("hi err is here 2")
			return err
		}
	}

	//天内跨小时,也写增量..
	if !spanday && util.JudgeInSameHour(uc.LastuploadTime, time.Now().Unix()) == false {

		fmt.Println("in tian kua", uc.LastuploadTime, time.Now().Unix())

		var currentsteps, formersteps int
		if rsteps, ok := Map.Get(uc.Uid); ok {
			currentsteps = rsteps.(int)
		} else {
			return errors.New("map中没有user:" + strconv.Itoa(uc.Uid))
		}

		if fsteps, ok := MapOld.Get(uc.Uid); ok {
			formersteps = fsteps.(int)
		} else {
			return errors.New("mapold中没有user:" + strconv.Itoa(uc.Uid))
		}

		//重新计算formersteps，需要从小时数据中解析出来所有的有数据的小时，然并加..
		if formersteps == 0 {

			hour := time.Unix(uc.LastuploadTime, 0).Hour()
			sqlStr := "select "
			for i := 0; i <= hour; i++ {

				column := "hour" + strconv.Itoa(i)
				sqlStr += "SUBSTRING_INDEX(" + column + ",',',1)+0 +"

			}

			//trim the last +
			sqlStr = sqlStr[0 : len(sqlStr)-1]
			sqlStr += "from wanbu_data_walkhour where userid = ? and walkdate = ?"

			rows, err := db.Query(sqlStr, uc.Uid, uc.Walkdays[0].Walkdate)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {

				err := rows.Scan(&formersteps)
				if err != nil {
					return err
				}
			}
			fmt.Println("重新计算后的formersteps ", formersteps)
		}

		increment := currentsteps - formersteps
		fmt.Println("increment is", increment)

		column := "hour" + strconv.Itoa(time.Now().Hour())

		newsteps := increment
		newstepwith := uc.Walkdays[0].Stepwidth
		newvalue := fmt.Sprintf("%d,%d,%d,%d,0,0", newsteps, newsteps*newstepwith, newsteps, newsteps*newstepwith)

		fmt.Println("天内跨小时写增量", newsteps, newstepwith, newvalue)

		//重新运算，更新DB
		us := "update wanbu_data_walkhour set " + column + "= '" + newvalue + "' where userid = ? and walkdate = ? "

		_, err := db.Exec(us, uc.Uid, uc.Walkdays[0].Walkdate)

		if err != nil {
			return err
		}

	}

	//小时内增量增长，拿到上次请求时的步数，跟此次的请求做比较，得到增量值,update时候算加和
	//!spanday && util.JudgeInSameHour(uc.LastuploadTime, time.Now().Unix()) == true
	//fmt.Println("span day", spanday)
	if !spanday && util.JudgeInSameHour(uc.LastuploadTime, time.Now().Unix()) == true {

		var currentsteps, formersteps int
		if rsteps, ok := Map.Get(uc.Uid); ok {
			currentsteps = rsteps.(int)
		} else {
			return errors.New("map中没有user:" + strconv.Itoa(uc.Uid))
		}

		if fsteps, ok := MapOld.Get(uc.Uid); ok {
			formersteps = fsteps.(int)
		} else {
			return errors.New("mapold中没有user:" + strconv.Itoa(uc.Uid))
		}

		increment := currentsteps - formersteps
		fmt.Println("increment is", increment)

		column := "hour" + strconv.Itoa(time.Now().Hour())

		//select SUBSTRING_INDEX(hour23,',',1)+0    from wanbu_data_walkhour where userid
		//数据库中找到当前小时数据，后期优化，同样放在map中，考虑程序崩溃后map的初始化?
		var oldvalue int
		qs := "select SUBSTRING_INDEX(" + column + ",',',1)+0 from wanbu_data_walkhour where userid = ? and walkdate = ?"

		rows, err0 := db.Query(qs, uc.Uid, uc.Walkdays[0].Walkdate)
		if err0 != nil {
			return err0
		}
		defer rows.Close()
		for rows.Next() {

			err := rows.Scan(&oldvalue)
			if err != nil {
				return err
			}
		}
		newsteps := oldvalue + increment
		newstepwith := uc.Walkdays[0].Stepwidth
		newvalue := fmt.Sprintf("%d,%d,%d,%d,0,0", newsteps, newsteps*newstepwith, newsteps, newsteps*newstepwith)

		fmt.Println("everything is new ", newsteps, newstepwith, newvalue)

		//重新运算，更新DB
		us := "update wanbu_data_walkhour set " + column + "= '" + newvalue + "' where userid = ? and walkdate = ? "

		//fmt.Println(us)
		_, err := db.Exec(us, uc.Uid, uc.Walkdays[0].Walkdate)

		if err != nil {
			//fmt.Println("hi err is here 3")
			return err
		}
	}

	return nil
}

func InsertWalkDay(db *sql.DB, uc *User_walkdays_struct) error {

	//if DaysDiff
	//IF(stepnumber < VALUES(stepnumber),VALUES(exerciseamount), exerciseamount)
	sqlStr := `
	   INSERT INTO wanbu_data_walkday (userid, walkdate, timestamp, servertime, deviceserial, stepwidth, weight, goalstepnum, stepnumber, walkdistance, walktime, calorieconsumed, fatconsumed, exerciseamount, heartvalue, timezonenum, timezone, manualflag, zmflag, faststepnum, dataflag) 
	   VALUES`

	vals := []interface{}{}

	for _, uds := range uc.Walkdays {
		sqlStr += "(?,?,UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),?,?,?,?,?,?,?,?,?,?,0,8,0,0,0,0,0),"

		vals = append(vals, uc.Uid, uds.Walkdate, miuserial, uds.Stepwidth, uds.Weight, uds.Goalstepnum, uds.Stepnumber, uds.Walkdistance, uds.Walktime, uds.Calorieconsumed, uds.Fatconsumed, uds.Exerciseamount)
	}

	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-1]

	/*
		sqlStr += `ON DUPLICATE KEY UPDATE timestamp = VALUES(timestamp),servertime = VALUES(servertime),stepwidth = VALUES(stepwidth),weight = VALUES(weight),goalstepnum = VALUES(goalstepnum),stepnumber = VALUES(stepnumber),walkdistance = VALUES(walkdistance),walktime = VALUES(walktime),calorieconsumed = VALUES(calorieconsumed),fatconsumed = VALUES(fatconsumed),exerciseamount = VALUES(exerciseamount)`
	*/

	/*
		sqlStr += `ON DUPLICATE KEY UPDATE timestamp =  IF(stepnumber < VALUES(stepnumber),VALUES(timestamp), timestamp),servertime = IF(stepnumber < VALUES(stepnumber),VALUES(servertime), servertime),stepwidth = IF(stepnumber < VALUES(stepnumber),VALUES(stepwidth), stepwidth),weight = IF(stepnumber < VALUES(stepnumber),VALUES(weight), weight),goalstepnum = IF(stepnumber < VALUES(stepnumber),VALUES(goalstepnum), goalstepnum),stepnumber = IF(stepnumber < VALUES(stepnumber),VALUES(stepnumber), stepnumber),walkdistance = IF(stepnumber < VALUES(stepnumber),VALUES(walkdistance), walkdistance),walktime = IF(stepnumber < VALUES(stepnumber),VALUES(walktime), walktime),calorieconsumed = IF(stepnumber < VALUES(stepnumber),VALUES(calorieconsumed), calorieconsumed),fatconsumed = IF(stepnumber < VALUES(stepnumber),VALUES(fatconsumed), fatconsumed),exerciseamount = IF(stepnumber < VALUES(stepnumber),VALUES(exerciseamount), exerciseamount)`
	*/

	sqlStr += `
		ON DUPLICATE KEY UPDATE timestamp =  IF(stepnumber < VALUES(stepnumber),VALUES(timestamp), timestamp),servertime = IF(stepnumber < VALUES(stepnumber),VALUES(servertime), servertime),stepwidth = IF(stepnumber < VALUES(stepnumber),VALUES(stepwidth), stepwidth),weight = IF(stepnumber < VALUES(stepnumber),VALUES(weight), weight),goalstepnum = IF(stepnumber < VALUES(stepnumber),VALUES(goalstepnum), goalstepnum),walkdistance = IF(stepnumber < VALUES(stepnumber),VALUES(walkdistance),walkdistance),walktime = IF(stepnumber < VALUES(stepnumber),VALUES(walktime), walktime),calorieconsumed = IF(stepnumber < VALUES(stepnumber),VALUES(calorieconsumed), calorieconsumed),fatconsumed = IF(stepnumber < VALUES(stepnumber),VALUES(fatconsumed), fatconsumed),exerciseamount = IF(stepnumber < VALUES(stepnumber),VALUES(exerciseamount), exerciseamount),deviceserial=IF(stepnumber < VALUES(stepnumber),VALUES(deviceserial),deviceserial),stepnumber = IF(stepnumber < VALUES(stepnumber),VALUES(stepnumber), stepnumber)`

	//format all vals at once
	_, err := db.Exec(sqlStr, vals...)

	if err != nil {
		return err
	}

	return nil
}

//填充AnswerData中stepwidth,weight,stepgoal的数据，从wanbu_data_user表中获得
func SetAnswerData(db *sql.DB, ad *AnswerData, uid int) error {

	qs := "select stepwidth,weight,stepgoal from wanbu_data_user where userid=?"
	rows, err := db.Query(qs, uid)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		/*
		 Stepwidth       int
		 Weight          int
		 Goalstepnum     int
		*/
		err := rows.Scan(&ad.Stepwidth, &ad.Weight, &ad.Goalstepnum)
		if err != nil {
			return err
		}
	}

	return nil
}

func ModifyLastuploadtime(db *sql.DB, m *User_walkdays_struct) error {

	us := "update wanbu_data_userdevice set lastuploadtime = UNIX_TIMESTAMP() where userid=?"
	_, err := db.Exec(us, m.Uid)

	if err != nil {
		return err
	}
	return nil
}

func ModifyStatus(db *sql.DB, uid int) error {

	us := "update wanbu_mi_sync set status = 1 where userid=?"
	_, err := db.Exec(us, uid)

	if err != nil {
		return err
	}
	return nil
}

//检查wanbu_mi_sync表中的这些人，有哪些已经不再使用小米设备了，更新flag字段
func ModifyPerson(db *sql.DB) error {

	us := "update wanbu_mi_sync ws , wanbu_data_userdevice wu set flag = 1  where ws.userid = wu.userid and wu.deviceserial !=?"

	_, err := db.Exec(us, miuserial)

	if err != nil {
		return err
	}
	return nil
}

func GetAllPerson1(db *sql.DB) ([]*Miu, error) {

	res := []*Miu{}

	//一次性获取需要更新数据的人，找到授权码未过期并且当前绑定设备为小秘手环的人
	//qs := "select userid,appid,accesstoken,mackey from  wanbu_mi_sync where flag=0 and status =0 "
	qs := "select ws.userid,ws.appid,ws.accesstoken,ws.mackey,from_unixtime(wu.lastuploadtime,'%Y-%m-%d'),wu.lastuploadtime from  wanbu_mi_sync ws,wanbu_data_userdevice wu where ws.flag=0 and ws.status =0 and ws.userid = wu.userid"
	rows, err := db.Query(qs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var t int64 = time.Now().Unix()
	//截止到当前时间
	var s string = time.Unix(t, 0).Format("2006-01-02")
	for rows.Next() {

		re := &Miu{}
		re.MiuPrepare()
		err := rows.Scan(&re.Userid, &re.Appid, &re.Access_token, &re.Mac_key, &re.Fromdate, &re.LastuploadTime)
		if err != nil {
			return nil, err
		}

		re.Todate = s
		res = append(res, re)

	}
	//fmt.Println("GetAllPerson...", res)
	return res, nil
}

func GetAllPerson(db *sql.DB) ([]*Miu, error) {

	res := []*Miu{}

	//一次性获取需要更新数据的人，找到授权码未过期并且当前绑定设备为小秘手环的人
	qs := "select userid,appid,accesstoken,mackey from  wanbu_mi_sync where flag=0 and status =0 "

	rows, err := db.Query(qs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {

		re := &Miu{}
		re.MiuPrepare()
		err := rows.Scan(&re.Userid, &re.Appid, &re.Access_token, &re.Mac_key)
		if err != nil {
			return nil, err
		}

		err0 := GetDate(re, db)
		if err0 != nil {
			return nil, err0
		}
		res = append(res, re)

	}
	//fmt.Println("GetAllPerson...", res)
	return res, nil
}

func GetDate(re *Miu, db *sql.DB) error {

	//格式化时间，测试过如果lastuploadtime为0自动赋值为1970年
	qs := "select from_unixtime(lastuploadtime,'%Y-%m-%d'),lastuploadtime from wanbu_data_userdevice where userid=?"
	rows, err := db.Query(qs, re.Userid)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&re.Fromdate, &re.LastuploadTime)
		if err != nil {
			return err
		}
	}
	//re.Fromdate = "2016-05-26"
	var t int64 = time.Now().Unix()
	var s string = time.Unix(t, 0).Format("2006-01-02")
	//截止到当前时间
	re.Todate = s
	return nil
}
