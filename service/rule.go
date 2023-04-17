package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
)

var (
	TotalCards   = "A23456789TJQK"
	Pokers       = make(map[string]*Combination, 16384)
	TypeToPokers = make(map[string][]*Combination, 38)
)

const (
	single             = "single"
	pair               = "pair"
	triplet            = "triplet"
	bomb               = "bomb"
	rocket             = "rocket"
	seq_single         = "seq_single"
	seq_pair           = "seq_pair"
	seq_triplet        = "seq_triplet"
	triplet_single     = "triplet_single"
	triplet_pair       = "triplet_pair"
	seq_triplet_single = "seq_triplet_single"
	seq_triplet_pair   = "seq_triplet_pair"
	bomb_single        = "bomb_single"
	bomb_pair          = "bomb_pair"
)

type Combination struct {
	Type  string
	Score int
	Poker string
}

func init() {
	path := "static/rule.json"
	_, err := os.Stat(path)
	var rule = make(map[string][]string)
	if os.IsNotExist(err) {
		logs.Info(path, "is not exist, will generate")
		rule = generate(path)
	} else {
		content, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(content, &rule)
		if err != nil {
			fmt.Printf("json unmarsha1 err:%v \n", err)
			return
		}
	}

	for comboType, pokers := range rule {
		for score, poker := range pokers {
			cards := sortPokers(poker)
			p := &Combination{
				Type:  comboType,
				Score: score,
				Poker: cards,
			}
			Pokers[cards] = p
			TypeToPokers[comboType] = append(TypeToPokers[comboType], p)
		}
	}
}

// sortPokers Sort Poker with
func sortPokers(pokers string) (sortPokers string) {
	runeArr := make([]int, 0)
	for _, s := range pokers {
		runeArr = append(runeArr, int(s))
	}
	sort.Ints(runeArr)
	res := make([]byte, 0)
	for _, v := range runeArr {
		res = append(res, byte(v))
	}
	return string(res)
}

// isContains 出的牌是否在手牌中存在
func isContains(parent, child string) (result bool) {
	for _, childCard := range child {
		inHand := false
		for i, parentCard := range parent {
			if childCard == parentCard {
				inHand = true
				tmp := []byte(parent)
				copy(tmp[i:], tmp[i+1:])
				tmp = tmp[:len(tmp)-1]
				parent = string(tmp)
				break
			}
		}
		if !inHand {
			return
		}
	}
	return true
}

// toPokers 将牌编号转换为扑克牌
func toPokers(num []int) string {

	res := make([]byte, 0)
	for _, poker := range num {
		if poker == 52 {
			res = append(res, 'B')
		} else if poker == 53 {
			res = append(res, 'L')
		} else {
			res = append(res, TotalCards[poker%13])
		}
	}
	return string(res)
}

// toPoker 将牌转换为编号
func toPoker(card byte) (poker []int) {
	if card == 'B' {
		return []int{52}
	}
	if card == 'L' {
		return []int{53}
	}
	for i, c := range []byte(TotalCards) {
		if c == card {
			return []int{i, i + 13, i + 13*2, i + 13*3}
		}
	}
	return []int{54}
}

// 将机器人要出的牌转换为编号
func pokersInHand(num []int, findPokers string) (pokers []int) {
	var isInResPokers = func(poker int) bool {
		for _, p := range pokers {
			if p == poker {
				return true
			}
		}
		return false
	}

	for _, poker := range findPokers {
		poker := toPoker(byte(poker))
	out:
		for _, pItem := range poker {
			for _, n := range num {
				if pItem == n && !isInResPokers(n) {
					pokers = append(pokers, pItem)
					break out
				}
			}
		}
	}
	return
}

// 获得牌型和大小
func pokersValue(pokers string) (cardType string, score int) {
	if combination, ok := Pokers[sortPokers(pokers)]; ok {
		cardType = combination.Type
		score = combination.Score
	}
	return
}

// ComparePoker 比较牌大小,并返回是否翻倍
func ComparePoker(baseNum, comparedNum []int) (int, bool) {
	logs.Debug("comparedNum %v  %v", baseNum, comparedNum)
	if len(baseNum) == 0 || len(comparedNum) == 0 {
		if len(baseNum) == 0 && len(comparedNum) == 0 {
			return 0, false
		} else {
			if len(baseNum) != 0 {
				return -1, false
			} else {
				comparedType, _ := pokersValue(toPokers(comparedNum))
				if comparedType == rocket || comparedType == bomb {
					return 1, true
				}
				return 1, false
			}
		}
	}
	baseType, baseScore := pokersValue(toPokers(baseNum))
	comparedType, comparedScore := pokersValue(toPokers(comparedNum))
	logs.Debug("compare poker %v, %v, %v, %v", baseType, baseScore, comparedType, comparedScore)
	if baseType == comparedType {
		return comparedScore - baseScore, false
	}
	if comparedType == rocket {
		return 1, true
	}
	if baseType == rocket {
		return -1, false
	}
	if comparedType == bomb {
		return 1, true
	}
	return 0, false
}

// CardsAbove 查找手牌中是否有比被比较牌型大的牌
func CardsAbove(handsNum, lastShotNum []int) (aboveNum []int) {
	handCards := toPokers(handsNum)
	turnCards := toPokers(lastShotNum)
	cardType, cardScore := pokersValue(turnCards)
	logs.Debug("CardsAbove hands:%v ,lastShot:%v, handCards %v,cardType %v,turnCards %v",
		handsNum, lastShotNum, handCards, cardType, turnCards)
	if len(cardType) == 0 {
		return
	}
	for _, combination := range TypeToPokers[cardType] {
		if combination.Score > cardScore && isContains(handCards, combination.Poker) {
			aboveNum = pokersInHand(handsNum, combination.Poker)
			return
		}
	}
	if cardType != bomb && cardType != rocket {
		for _, combination := range TypeToPokers[bomb] {
			if isContains(handCards, combination.Poker) {
				aboveNum = pokersInHand(handsNum, combination.Poker)
				return
			}
		}
	} else if isContains(handCards, "LB") {
		aboveNum = pokersInHand(handsNum, "LB")
		return
	}
	return
}

// 根据sep，生成长度为num的顺子集合
func generateSeq(num int, seq []string) (res []string) {
	for i := range seq {
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

func generate(path string) map[string][]string {
	rule := map[string][]string{}
	rule[single] = []string{}
	rule[pair] = []string{}
	rule[triplet] = []string{}
	rule[bomb] = []string{}
	for _, c := range TotalCards {
		card := string(c)
		rule[single] = append(rule[single], card)
		rule[pair] = append(rule[pair], card+card)
		rule[triplet] = append(rule[triplet], card+card+card)
		rule[bomb] = append(rule[bomb], card+card+card+card)
	}
	// 生成单顺  34567
	for _, num := range []int{5, 6, 7, 8, 9, 10, 11, 12} {
		rule[seq_single+strconv.Itoa(num)] = generateSeq(num, rule[single])
	}
	// 生成连对 334455
	for _, num := range []int{3, 4, 5, 6, 7, 8, 9, 10} {
		rule[seq_pair+strconv.Itoa(num)] = generateSeq(num, rule[pair])
	}
	// 飞机不带翅膀  333444
	for _, num := range []int{2, 3, 4, 5, 6} {
		rule[seq_triplet+strconv.Itoa(num)] = generateSeq(num, rule[triplet])
	}

	rule[single] = append(rule[single], "L")
	rule[single] = append(rule[single], "B")
	rule[rocket] = append(rule[rocket], "LB")

	rule[triplet_single] = make([]string, 0)
	rule[triplet_pair] = make([]string, 0)

	for _, t := range rule[triplet] {
		for _, s := range rule[single] {
			if s[0] != t[0] {
				rule[triplet_single] = append(rule[triplet_single], t+s)
			}
		}
		for _, p := range rule[pair] {
			if p[0] != t[0] {
				rule[triplet_pair] = append(rule[triplet_pair], t+p)
			}
		}
	}
	for _, num := range []int{2, 3, 4, 5} {
		seqTripletSingle := []string(nil)
		seqTripletPair := []string(nil)
		for _, seqTriplet := range rule[seq_triplet+strconv.Itoa(num)] {
			seq := make([]string, len(rule[single]))
			copy(seq, rule[single])
			for i := 0; i < len(seqTriplet); i = i + 3 {
				for k, v := range seq {
					if v[0] == seqTriplet[i] {
						copy(seq[k:], seq[k+1:])
						seq = seq[:len(seq)-1]
						break
					}
				}
			}
			for _, singleCombination := range combination(seq, len(seqTriplet)/3) {
				seqTripletSingle = append(seqTripletSingle, seqTriplet+singleCombination)
				var hasJoker bool
				for _, single := range singleCombination {
					if single == 'L' || single == 'B' {
						hasJoker = true
					}
				}
				if !hasJoker {
					seqTripletPair = append(seqTripletPair, seqTriplet+singleCombination+singleCombination)
				}
			}
		}
		rule[seq_triplet_single+strconv.Itoa(num)] = seqTripletSingle
		rule[seq_triplet_pair+strconv.Itoa(num)] = seqTripletPair
	}

	rule[bomb_single] = []string(nil)
	rule[bomb_pair] = []string(nil)
	for _, b := range rule[bomb] {
		seq := make([]string, len(rule[single]))
		copy(seq, rule[single])
		for i, single := range seq {
			if single[0] == b[0] {
				copy(seq[i:], seq[i+1:])
				seq = seq[:len(seq)-1]
			}
		}
		for _, comb := range combination(seq, 2) {
			rule[bomb_single] = append(rule[bomb_single], b+comb)
			if comb[0] != 'L' && comb[0] != 'B' && comb[1] != 'L' && comb[1] != 'B' {
				rule[bomb_pair] = append(rule[bomb_pair], b+comb+comb)
			}
		}
	}

	res, err := json.Marshal(rule)
	if err == nil {
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
