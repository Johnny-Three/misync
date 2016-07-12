package structure

import (
	"fmt"
	"time"
	"wbproject/miusync/util"
)

var Map *util.BeeMap

//存储上次请求数据时，每个用户的步数，用作小数内增量步数计算
//MapOld与Map有相同个数的元素，key相同，Value不同，只有在步数变化的时候才会写入MapOld
var MapOld *util.BeeMap

func AssginHourData(hour *HourData, uds *AnswerData) {

	switch time.Now().Hour() {

	case 0:
		hour.H0 = uds.GetHourString()
	case 1:
		hour.H1 = uds.GetHourString()
	case 2:
		hour.H2 = uds.GetHourString()
	case 3:
		hour.H3 = uds.GetHourString()
	case 4:
		hour.H4 = uds.GetHourString()
	case 5:
		hour.H5 = uds.GetHourString()
	case 6:
		hour.H6 = uds.GetHourString()
	case 7:
		hour.H7 = uds.GetHourString()
	case 8:
		hour.H8 = uds.GetHourString()
	case 9:
		hour.H9 = uds.GetHourString()
	case 10:
		hour.H10 = uds.GetHourString()
	case 11:
		hour.H11 = uds.GetHourString()
	case 12:
		hour.H12 = uds.GetHourString()
	case 13:
		hour.H13 = uds.GetHourString()
	case 14:
		hour.H14 = uds.GetHourString()
	case 15:
		hour.H15 = uds.GetHourString()
	case 16:
		hour.H16 = uds.GetHourString()
	case 17:
		hour.H17 = uds.GetHourString()
	case 18:
		hour.H18 = uds.GetHourString()
	case 19:
		hour.H19 = uds.GetHourString()
	case 20:
		hour.H20 = uds.GetHourString()
	case 21:
		hour.H21 = uds.GetHourString()
	case 22:
		hour.H22 = uds.GetHourString()
	case 23:
		hour.H23 = uds.GetHourString()

	}
}

type Miu struct {
	Userid          int
	LastuploadTime  int64
	Appid           string `json:"appid"`
	Third_appid     string `json:"third_appid"`
	Third_appsecret string `json:"third_appsecret"`
	Mac_key         string `json:"mac_key"`
	Call_id         string `json:"call_id"`
	Access_token    string `json:"access_token"`
	Fromdate        string `json:"fromdate"`
	Todate          string `json:"todate"`
	V               string `json:"v"`
	L               string `json:"l"`
}

func (t *Miu) MiuPrepare() {

	t.Third_appid = "1464244765"
	t.Third_appsecret = "b5423ca9d2deff0ea76bd88871d293f3"
	t.Call_id = "1464244765"
	t.V = "1.0.0"
	t.L = "english"
}

type User_walkdays_struct struct {
	Uid            int
	LastuploadTime int64
	RecentTime     int64
	Walkdays       []AnswerData
}

type Reback struct {
	Userid         int
	LastuploadTime int64
	JsonCode       string
}

type AnswerData struct {
	Walkdate        int64
	Walkdistance    int
	Walktime        int
	Stepnumber      int
	Stepwidth       int
	Weight          int
	Goalstepnum     int
	Calorieconsumed float64
	Fatconsumed     float64
	Exerciseamount  float64
}

func (t *AnswerData) GetHourString() string {
	s := fmt.Sprintf("%d,%d,%d,%d,0,0", t.Stepnumber, t.Stepnumber*t.Stepwidth, t.Stepnumber, t.Stepnumber*t.Stepwidth)
	return s
}

type DayData struct {
	Userid          int64
	Walkdate        int64
	Timestamp       int64
	Servertime      int64
	Deviceserial    string
	Stepwidth       int
	Weight          int
	Goalstepnum     int
	Stepnumber      int
	Walkdistance    int64
	Walktime        int
	Calorieconsumed float64
	Fatconsumed     float64
	Exerciseamount  float64
}

type HourData struct {
	H0  string
	H1  string
	H2  string
	H3  string
	H4  string
	H5  string
	H6  string
	H7  string
	H8  string
	H9  string
	H10 string
	H11 string
	H12 string
	H13 string
	H14 string
	H15 string
	H16 string
	H17 string
	H18 string
	H19 string
	H20 string
	H21 string
	H22 string
	H23 string
}

func (t *HourData) Init() {

	t.H0 = "0,0,0,0,0,0"
	t.H1 = "0,0,0,0,0,0"
	t.H2 = "0,0,0,0,0,0"
	t.H3 = "0,0,0,0,0,0"
	t.H4 = "0,0,0,0,0,0"
	t.H5 = "0,0,0,0,0,0"
	t.H6 = "0,0,0,0,0,0"
	t.H7 = "0,0,0,0,0,0"
	t.H8 = "0,0,0,0,0,0"
	t.H9 = "0,0,0,0,0,0"
	t.H10 = "0,0,0,0,0,0"
	t.H11 = "0,0,0,0,0,0"
	t.H12 = "0,0,0,0,0,0"
	t.H13 = "0,0,0,0,0,0"
	t.H14 = "0,0,0,0,0,0"
	t.H15 = "0,0,0,0,0,0"
	t.H16 = "0,0,0,0,0,0"
	t.H17 = "0,0,0,0,0,0"
	t.H18 = "0,0,0,0,0,0"
	t.H19 = "0,0,0,0,0,0"
	t.H20 = "0,0,0,0,0,0"
	t.H21 = "0,0,0,0,0,0"
	t.H22 = "0,0,0,0,0,0"
	t.H23 = "0,0,0,0,0,0"
}
