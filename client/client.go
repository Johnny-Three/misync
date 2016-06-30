package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	. "wbproject/miusync/logs"
	. "wbproject/miusync/structure"
)

var Post_request_chan chan Reback

func init() {

	Post_request_chan = make(chan Reback, 1024)
}

func Post(r *Miu) {

	url := "https://hmservice.mi-ae.com.cn/user/summary/getData"

	re := fmt.Sprintf("%s?appid=%s&third_appid=%s&third_appsecret=%s&access_token=%s&mac_key=%s&call_id=%s&fromdate=%s&todate=%s&v=%s&l=%s", url, r.Appid, r.Third_appid, r.Third_appsecret, r.Access_token, r.Mac_key, r.Call_id, r.Fromdate, r.Todate, r.V, r.L)

	//fmt.Printf("userid【%d】,request【%s】", r.Userid, re)

	resp, err := http.Get(re)

	defer resp.Body.Close()
	if err != nil {
		Logger.Critical(err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		var re Reback
		re.Userid = r.Userid
		re.LastuploadTime = r.LastuploadTime
		re.JsonCode = string(body)
		Post_request_chan <- re
	}
}
