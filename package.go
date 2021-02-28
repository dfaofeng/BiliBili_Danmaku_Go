package danmaku

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io/ioutil"
	"log"
)
func packageHeadGen(body string, protocol uint16, operationCode uint32, sequence uint32) []byte {
	var headBuffer bytes.Buffer
	var bodyLength uint32 = uint32(len(body))
	var headLength uint16 = uint16(16)
	var packageLength uint32 = bodyLength + uint32(headLength)
	var bpackageLength []byte = make([]byte, 4)
	var bheadLength []byte = make([]byte, 2)
	var bprotocol []byte = make([]byte, 2)
	var boperationCode []byte = make([]byte, 4)
	var bsequence []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(bpackageLength, packageLength)
	headBuffer.Write(bpackageLength)
	binary.BigEndian.PutUint16(bheadLength, headLength)
	headBuffer.Write(bheadLength)
	binary.BigEndian.PutUint16(bprotocol, protocol)
	headBuffer.Write(bprotocol)
	binary.BigEndian.PutUint32(boperationCode, operationCode)
	headBuffer.Write(boperationCode)
	binary.BigEndian.PutUint32(bsequence, sequence)
	headBuffer.Write(bsequence)
	return headBuffer.Bytes()
}

func packageBodyGen(body string) []byte {
	return []byte(body)
}
//对订阅信息头部进行处理
func authAndHeartBeatPackageGen(authBody string) (authPackage []byte, heartBeatPackage []byte) {
	var authPackageBuffer bytes.Buffer
	authPackageBuffer.Write(packageHeadGen(authBody, 1, 7, 1))
	authPackageBuffer.Write(packageBodyGen(authBody))
	var heartBeatPackageBuffer bytes.Buffer
	heartBeatBody := "[object Object]"
	heartBeatPackageBuffer.Write(packageHeadGen(heartBeatBody, 1, 2, 1))
	heartBeatPackageBuffer.Write(packageBodyGen(heartBeatBody))
	return authPackageBuffer.Bytes(), heartBeatPackageBuffer.Bytes()
}
//处理头部信息
func wsPackageRead(msg []byte) []byte {
	//获得大端uint32的长度
	packageLength := binary.BigEndian.Uint32(msg[:4])
	if int(packageLength) != len(msg) {
		log.Println("The package is invalid")
	}
	//获得小端uint16的长度
	headLength := binary.BigEndian.Uint16(msg[4:6])
	content := msg[headLength:]
	return content
}
//根据协议头分离信息,并且返回切片数据
func zlibPackageRead(msg []byte) [][]byte {
	var contents [][]byte
	for {
		jsonPackageLength := binary.BigEndian.Uint32(msg[:4])
		jsonHeadLength := binary.BigEndian.Uint16(msg[4:6])
		contents = append(contents, msg[jsonHeadLength:jsonPackageLength])
		msg = msg[jsonPackageLength:]
		if len(msg) == 0 {
			break
		}
	}
	return contents
}
//zlib解压解压
func unzlib(msg []byte) []byte {
	b := bytes.NewReader(msg)
	r, err := zlib.NewReader(b)
	if err != nil {
		panic(err)
	}
	danmaku, _ := ioutil.ReadAll(r)
	return danmaku
}
