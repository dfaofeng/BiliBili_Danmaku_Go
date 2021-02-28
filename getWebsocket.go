package danmaku

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type RoomInfo struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Message string `json:"message"`
	Data struct {
		RoomID int `json:"room_id"`
		ShortID int `json:"short_id"`
		UID int `json:"uid"`
		NeedP2P int `json:"need_p2p"`
		IsHidden bool `json:"is_hidden"`
		IsLocked bool `json:"is_locked"`
		IsPortrait bool `json:"is_portrait"`
		LiveStatus int `json:"live_status"`
		HiddenTill int `json:"hidden_till"`
		LockTill int `json:"lock_till"`
		Encrypted bool `json:"encrypted"`
		PwdVerified bool `json:"pwd_verified"`
		LiveTime int64 `json:"live_time"`
		RoomShield int `json:"room_shield"`
		IsSp int `json:"is_sp"`
		SpecialType int `json:"special_type"`
	} `json:"data"`
}
type DanmuInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Group            string  `json:"group"`
		BusinessID       int     `json:"business_id"`
		RefreshRowFactor float64 `json:"refresh_row_factor"`
		RefreshRate      int     `json:"refresh_rate"`
		MaxDelay         int     `json:"max_delay"`
		Token            string  `json:"token"`
		HostList         []struct {
			Host    string `json:"host"`
			Port    int    `json:"port"`
			WssPort int    `json:"wss_port"`
			WsPort  int    `json:"ws_port"`
		} `json:"host_list"`
	} `json:"data"`
}

//主函数
func InitRoom(room int) chan []byte {
	//短号转换
	roominfo, err := getRoom(room)
	if err != nil {
		return nil
	}
	//获取真实房间号
	roomId := strconv.Itoa(roominfo.Data.RoomID)
	//获取弹幕服务器地址
	danmakuInfo,err :=apiDanmuInfoRequest(roomId)
	//ws服务器连接,处理
	return websocketConnection( roomId, danmakuInfo)
}
//json序列化
func jsonInfoDecode(body []byte, v interface{}) {
	err := json.Unmarshal(body, v)
	if err != nil {
		log.Fatalln("[ERROR]", err)
	}
}
//获取弹幕服务端地址
func apiDanmuInfoRequest(roomID string) (danmakuInfo DanmuInfo, err error) {
	log.Println("[INFO] 正在获取弹幕源信息")
	client := &http.Client{}
	apiURL := "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=" + roomID + "&type=0"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] 无法获取弹幕源信息")
		log.Println("[Detail]", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	getH5InfoByRoom:="https://api.live.bilibili.com/xlive/web-room/v1/index/getH5InfoByRoom?room_id="+roomID
	reqs, err := http.NewRequest("GET", getH5InfoByRoom, nil)
	if err != nil {
		log.Println(err)
	}
	resps, err := client.Do(reqs)
	if err != nil {
		log.Println("[ERROR] 无法获取直播信息")
		log.Println("[Detail]", err)
	}
	defer resps.Body.Close()
	bodys, _ := ioutil.ReadAll(resps.Body)
	lastJson :=gjson.ParseBytes(bodys)
	uname :=lastJson.Get("data.anchor_info.base_info.uname").String()
	title :=lastJson.Get("data.room_info.title").String()
	area_name :=lastJson.Get("data.room_info.area_name").String()
	log.Printf("[INFO] 主播昵称:%s\n", uname)
	log.Printf("[INFO] 直播标题:%s\n", title)
	log.Printf("[INFO] 直播分区:%s\n", area_name)
	if resp.StatusCode != 200 {
		log.Println(resp.Status)
		log.Println(string(body))
	} else {
		jsonInfoDecode(body, &danmakuInfo)
		return danmakuInfo, nil
	}

	return danmakuInfo, err
}
//获取真实房间号
func getRoom(roomId int) (roominfo RoomInfo,err error) {
	log.Println("短号转换中...")
	client :=&http.Client{}
	apiUrl :="https://api.live.bilibili.com/room/v1/Room/room_init?id="+ strconv.Itoa(roomId)
	req,err :=http.NewRequest("GET",apiUrl,nil)
	if err != nil {
		log.Println(err)
	}
	resp,err :=client.Do(req)
	if err != nil {
		log.Println("获取直播间信息失败")
		log.Println("",err)
	}
	defer resp.Body.Close()
	body,_:=ioutil.ReadAll(resp.Body)
	if resp.StatusCode !=200{
		log.Panicln(resp.StatusCode)
	}else {
		jsonInfoDecode(body,&roominfo)
	}
	return
}
