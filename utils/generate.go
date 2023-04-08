package utils

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
)

// 生成连续num个的单牌的顺子
func generateSeq(num int, seq []string) (res []string) {
	for i, _ := range seq {
		if i+num > 12 {
			break
		}
		var sec string
		for j := i; j < i+num; j++ {
			sec += seq[j]
		}
		res = append(res, sec)
	}
	return
}

// 生成num个不同单的组合
func combination(seq []string, num int) (comb []string) {
	if num == 0 {
		panic("generate err , combination count can not be 0")
	}
	if len(seq) < num {
		logs.Error("seq: %v,num:%d", seq, num)
		return
		//panic("generate err , seq length less than num")
	}
	if num == 1 {
		return seq
	}
	if len(seq) == num {
		allSingle := ""
		for _, single := range seq {
			allSingle += single
		}
		return []string{allSingle}
	}
	noFirst := combination(seq[1:], num)
	hasFirst := []string(nil)
	for _, comb := range combination(seq[1:], num-1) {
		hasFirst = append(hasFirst, string(seq[0])+comb)
	}
	comb = append(comb, noFirst...)
	comb = append(comb, hasFirst...)
	return
}

func Generate(path string) map[string][]string {
	cards := "3456789TJQKA2"
	rule := map[string][]string{}
	rule["single"] = []string{}
	rule["pair"] = []string{}
	rule["trio"] = []string{}
	rule["bomb"] = []string{}
	for _, c := range cards {
		card := string(c)
		rule["single"] = append(rule["single"], card)
		rule["pair"] = append(rule["pair"], card+card)
		rule["trio"] = append(rule["trio"], card+card+card)
		rule["bomb"] = append(rule["bomb"], card+card+card+card)
	}
	for _, num := range []int{5, 6, 7, 8, 9, 10, 11, 12} {
		rule["seq_single"+strconv.Itoa(num)] = generateSeq(num, rule["single"])
	}
	for _, num := range []int{3, 4, 5, 6, 7, 8, 9, 10} {
		rule["seq_pair"+strconv.Itoa(num)] = generateSeq(num, rule["pair"])
	}
	for _, num := range []int{2, 3, 4, 5, 6} {
		rule["seq_trio"+strconv.Itoa(num)] = generateSeq(num, rule["trio"])
	}
	rule["single"] = append(rule["single"], "L")
	rule["single"] = append(rule["single"], "B")
	rule["rocket"] = append(rule["rocket"], "BL")

	rule["trio_single"] = make([]string, 0)
	rule["trio_pair"] = make([]string, 0)

	for _, t := range rule["trio"] {
		for _, s := range rule["single"] {
			if s[0] != t[0] {
				rule["trio_single"] = append(rule["trio_single"], t+s)
			}
		}
		for _, p := range rule["pair"] {
			if p[0] != t[0] {
				rule["trio_pair"] = append(rule["trio_pair"], t+p)
			}
		}
	}
	for _, num := range []int{2, 3, 4, 5} {
		seqTrioSingle := []string(nil)
		seqTrioPair := []string(nil)
		for _, seqTrio := range rule["seq_trio"+strconv.Itoa(num)] {
			seq := make([]string, len(rule["single"]))
			copy(seq, rule["single"])
			for i := 0; i < len(seqTrio); i = i + 3 {
				for k, v := range seq {
					if v[0] == seqTrio[i] {
						copy(seq[k:], seq[k+1:])
						seq = seq[:len(seq)-1]
						break
					}
				}
			}
			for _, singleCombination := range combination(seq, len(seqTrio)/3) {
				seqTrioSingle = append(seqTrioSingle, seqTrio+singleCombination)
				var hasJoker bool
				for _, single := range singleCombination {
					if single == 'L' || single == 'B' {
						hasJoker = true
					}
				}
				if !hasJoker {
					seqTrioPair = append(seqTrioPair, seqTrio+singleCombination+singleCombination)
				}
			}
		}
		rule["seq_trio_single"+strconv.Itoa(num)] = seqTrioSingle
		rule["seq_trio_pair"+strconv.Itoa(num)] = seqTrioPair
	}

	rule["bomb_single"] = []string(nil)
	rule["bomb_pair"] = []string(nil)
	for _, b := range rule["bomb"] {
		seq := make([]string, len(rule["single"]))
		copy(seq, rule["single"])
		for i, single := range seq {
			if single[0] == b[0] {
				copy(seq[i:], seq[i+1:])
				seq = seq[:len(seq)-1]
			}
		}
		for _, comb := range combination(seq, 2) {
			rule["bomb_single"] = append(rule["bomb_single"], b+comb)
			if comb[0] != 'L' && comb[0] != 'B' && comb[1] != 'L' && comb[1] != 'B' {
				rule["bomb_pair"] = append(rule["bomb_pair"], b+comb+comb)
			}
		}
	}

	res, err := json.Marshal(rule)
	if err != nil {
		panic("json marsha1 RULE err :" + err.Error())
	}
	file, err := os.Create(path)
	defer func() {
		err = file.Close()
		if err != nil {
			logs.Error("generate err: %v", err)
		}
	}()
	if err != nil {
		panic("create rule.json err:" + err.Error())
	}
	_, err = file.Write(res)
	if err != nil {
		panic("create rule.json err:" + err.Error())
	}
	return rule
}
