package service

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/beego/beego/v2/core/logs"
)

var (
	Pokers       = make(map[string]*Combination, 16384)
	TypeToPokers = make(map[string][]*Combination, 38)
)

type Combination struct {
	Type  string
	Score int
	Poker string
}

func init() {
	path := "static/rule.json"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		logs.Info(path, "is not exist, will generate")
		generate(path)
	}
	// file, err := os.Open(path)
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()

	// var content []byte
	// for {
	// 	buf := make([]byte, 1024)
	// 	readNum, err := file.Read(buf)
	// 	if err != nil && err != io.EOF {
	// 		panic(err)
	// 	}
	// 	for i := 0; i < readNum; i++ {
	// 		content = append(content, buf[i])
	// 	}
	// 	if 0 == readNum {
	// 		break
	// 	}
	// }

	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var rule = make(map[string][]string)
	err = json.Unmarshal(content, &rule)
	if err != nil {
		fmt.Printf("json unmarsha1 err:%v \n", err)
		return
	}
	for pokerType, pokers := range rule {
		for score, poker := range pokers {
			cards := SortStr(poker)
			p := &Combination{
				Type:  pokerType,
				Score: score,
				Poker: cards,
			}
			Pokers[cards] = p
			TypeToPokers[pokerType] = append(TypeToPokers[pokerType], p)
		}
	}
}
