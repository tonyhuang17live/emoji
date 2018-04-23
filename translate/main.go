package main

import (
	"fmt"

	"github.com/17media/emoji"
)

func main() {
	if err := emoji.BuildTable("../data"); err != nil {
		fmt.Println(err)
		return
	}

	if err := emoji.WriteToGo("../emojiMap.go"); err != nil {
		fmt.Println(err)
		return
	}
}
