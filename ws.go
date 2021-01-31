package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"time"
)
type ReturnCode struct {
	Code int `json:"code"`
}
//获取连接弹幕服务器所需要的url和token
func danmakuInfoExtract(danmakuInfo DanmuInfo) (hostURL []string, token string) {
	token = danmakuInfo.Data.Token
	hostURL = make([]string, 3)
	for i, src := range danmakuInfo.Data.HostList {
		hostURL[i] = "wss://" + src.Host + "/sub"
	}
	return hostURL, token
}
//心跳保持
func keepHeartBeat(ws *websocket.Conn, heartBeatPackage []byte) {
	ws.WriteMessage(websocket.BinaryMessage, heartBeatPackage)
}
//设置头部
func initWebsocket(hostURL []string) (*websocket.Conn, error) {
	var ws *websocket.Conn
	var err error
	for i := range hostURL {
		req, err := http.NewRequest("GET", hostURL[i], nil)
		req.Header.Add("origin", "https://live.bilibili.com")
		ws, _, err = websocket.DefaultDialer.Dial(hostURL[i], req.Header)
		if err != nil {
			log.Println(err)
			log.Println("[ERROR] 线路", i+1, "连接失败")
		} else {
			log.Println("[INFO] 线路", i+1, "连接成功")
			return ws, nil
		}
	}
	return ws, err
}
//处理返回的信息
func transferDanmaku( wsClient *websocket.Conn, danmakuChannel chan []byte, quit chan int) {
	for {
		select {
		case <-quit:
			log.Println("[INFO] 弹幕传输服务结束")
			//检测是否关闭
			close(danmakuChannel)
			return
		default:
			_, msg, err := wsClient.ReadMessage()
			//如果没有报错信息
			if err == nil {
				//取头长度
				rawContent := wsPackageRead(msg)

				if msg[11] == byte(5) && msg[7] == byte(2) {
					//根据头部长度解压byte
					unZippedContent := unzlib(rawContent)
					//去除头部信息，返回切片
					contents := zlibPackageRead(unZippedContent)
					//遍历切片
					for _, content := range contents {
						danmakuChannel <- content
					}
				} else if msg[11] == byte(5) && msg[7] == byte(0) {
					//返回头部信息
					danmakuChannel <- rawContent
				}
			} else {
				quit <- 1
				log.Println("[ERROR] 网络连接失败")
				close(danmakuChannel)
				return
			}
		}
	}
}
//首次连接服务器所需要的信息
func setupWebsocketConnection(roomID string, danmakuInfo DanmuInfo) ([]string, []byte, []byte) {
	//获取连接token和url
	hostURL, token := danmakuInfoExtract(danmakuInfo)
	//拼接相关信息
	authBody := `{"uid":0,"roomid":` + roomID + `,"protover":2,"platform":"web","clientver":"2.4.16","type":2,"key":"` + token + `"}`
	//对信息加以处理
	authPackage, heartBeatPackage := authAndHeartBeatPackageGen(authBody)
	return hostURL, authPackage, heartBeatPackage
}
//ws连接初始化
func websocketConnection(roomID string, danmakuInfo DanmuInfo) {
	restart := func(roomID string) {
		if err := recover(); err != nil {
			log.Println("[INFO] 正在尝试重连")
			time.Sleep(2000 * time.Millisecond)
			danmakuInfo, err := apiDanmuInfoRequest(roomID)
			if err != nil {
				log.Println("[INFO] 重连失败，请检查网络连接")
				return
			}
			websocketConnection( roomID, danmakuInfo)
		}
	}
	defer restart(roomID)
	//设置第一次连接ws服务器
	hostURL, authPackage, heartBeatPackage := setupWebsocketConnection(roomID, danmakuInfo)
	log.Println("[INFO] 正在连接到直播间", roomID)
	//设置头部
	wsClient, err := initWebsocket(hostURL)
	if err != nil {
		log.Println("[Panic] websocket client出错")
		log.Panicln("[Detail]", err)
	}
	//认证检查
	if err := wsClient.WriteMessage(websocket.TextMessage, authPackage); err != nil {
		log.Println("[Panic] websocket client出错 认证包问题")
		log.Panicln("[Detail]", err)
	}
	var msg = make([]byte, 512)
	//开始连接
	_, msg, err = wsClient.ReadMessage()
	if err != nil {
		log.Println("[Panic] websocket client出错")
		log.Panicln("[Detail]", err)
	}
	var returnCode ReturnCode
	//判断返回值
	if err := json.Unmarshal(msg[16:], &returnCode); err != nil {
		log.Println("[FATAL] 无法正常读取B站websocket信息")
		log.Fatalln("[Detail]", err)
	} else {
		if returnCode.Code == 0 {
			log.Println("[INFO] 成功连接到直播间")
		} else {
			log.Panicln("[Panic] Token无效")
		}
	}
	//发送心跳
	wsClient.WriteMessage(websocket.BinaryMessage, heartBeatPackage)
	danmakuChannel := make(chan []byte,1)
	quit := make(chan int)
	//重开线程,处理返回的信息
	go transferDanmaku( wsClient, danmakuChannel, quit)
	ticker := time.NewTicker(30 * time.Second)
	//go wsClient.ReadMessage()
	for  {
		select {
		case <-ticker.C:
			keepHeartBeat(wsClient, heartBeatPackage)
		case danmaku := <-danmakuChannel:
			GetCmd(danmaku)
		}
	}
}
//处理弹幕流
func GetCmd(Gjson []byte)  {
	lastJson :=gjson.ParseBytes(Gjson)
	switch lastJson.Get("cmd").String() {
	case "DANMU_MSG":
		fmt.Println(lastJson.Get("info.1").String())
	}
}
