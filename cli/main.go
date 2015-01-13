package main

import (
	"../xiami"
	"fmt"
)

func main() {
	music, _ := xiami.GetMusic(100)
	fmt.Println(music)
}
