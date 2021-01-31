package main

import (
	"fmt"
)

func main() {
	fmt.Printf("输入房间号: ")
	var room int
	fmt.Scanln(&room)
	initRoom(room)
}
