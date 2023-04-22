package common

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

// 小王:GL  大王:GB
var s2v = map[string]int{
	"3": 1, "4": 2, "5": 3, "6": 4, "7": 5, "8": 6, "9": 7, "T": 8, "J": 9, "Q": 10, "K": 11, "A": 12, "2": 13,
	"L": 14, "B": 15,
}
var v2s = map[int]string{
	1: "3", 2: "4", 3: "5", 4: "6", 5: "7", 6: "8", 7: "9", 8: "T", 9: "J", 10: "Q", 11: "K", 12: "A", 13: "2",
	14: "L", 15: "B",
}

const (
	TPass = iota
	// TSingle 单张
	TSingle
	// TPair 对子
	TPair
	// TTriple 三张 不带 例如 555
	TTriple
	// TTripletSingle 三带一
	TTripletSingle
	// TTripletPair 三代二
	TTripletPair
	// TSequenceSingle 单顺  	34567
	TSequenceSingle
	// TSSequencePair 双顺		334455
	TSSequencePair
	// TSequenceTriple 三顺
	TSequenceTriple
	// TSequenceTripleSingle 飞机带单	33344456
	TSequenceTripleSingle
	// TSequenceTriplePair 飞机代对	3334445566
	TSequenceTriplePair
	// TQuadSingle 四代一 	33334	333345
	TQuadSingle
	// TQuadPair 四代两队	33334455	333344
	TQuadPair
	// TBomb 炸弹		3333
	TBomb
	// TRocket 王炸(火箭)	GLGB
	TRocket
	// TWrong ERROR
	TWrong
)

const (
	MinSingle  = 5
	MinPairs   = 3
	MinTriples = 2
	TypeSingle = 1
	TypePair   = 2
	TypeTriple = 3
)

func FormatInputCards(v []string) ([]int, error) {
	var cards []int
	if len(v) == 0 || (len(v) == 1 && (v[0] == "Pass" || v[0] == "pass")) {
		return []int{}, nil //不要
	}
	for _, item := range v {
		if _, ok := s2v[strings.ToUpper(item)]; ok {
			cards = append(cards, s2v[item])
		} else {
			//若是没在字典表中如何处理？
			logs.Error(item, " not in s2v")
			return cards, fmt.Errorf(item + " not in s2v")
		}
	}
	sort.Ints(cards)
	return cards, nil
}

func FormatOutputCards(v []int) []string {
	var cards []string
	if len(v) == 0 {
		return []string{"Pass"} //不要
	}
	for _, item := range v {
		if _, ok := v2s[item]; ok {
			cards = append(cards, v2s[item])
		} else {
			logs.Error(item, " not in v2s")
		}
	}
	sort.Strings(cards)
	return cards
}

func ValidateCards(v []string) bool {
	for _, item := range v {
		if _, ok := s2v[item]; ok {
			return false
		}
	}
	return true
}

// GetRestCards 从手牌中剔除出的牌，获取剩余的手牌
func GetRestCards(cards *[]int, move *[]int) []int {
	cardsList := *cards
	if len(*move) == 0 {
		return cardsList
	}
	for _, card := range *move {
		findIt := false
		for i, it := range cardsList {
			if card == it {
				tempCardsList := append(cardsList[:i:i], cardsList[i+1:]...)
				cardsList = tempCardsList
				findIt = true
				break
			}
		}
		if !findIt {
			// 抛出错误
			logs.Info("cards", cards, " endgame", move)
			panic("endgame is not subset of cards")
		}
	}
	result := cardsList
	return result
}

// GetAnyN M >= N
type GetAnyN struct {
	Array       []int
	M           int
	N           int
	AllLists    [][]int
	FinalResult [][]int
}

func NewGetAnyN(array []int, m int, n int, allLists [][]int, finalResult [][]int) *GetAnyN {
	return &GetAnyN{array, m, n, allLists, finalResult}
}

func (o *GetAnyN) ListAllNOfM(m int, n int, result *[]int) {
	for i := m; i >= n; i-- {
		*result = append(*result, o.Array[i-1])
		if n > 1 {
			o.ListAllNOfM(i-1, n-1, result)
		} else {
			o.AllLists = append(o.AllLists, *result)
		}
		*result = (*result)[:len(*result)-1] //pop_back
	}
}

func (o *GetAnyN) RemoveDuplicated() {
	var duplicated []int
	size := len(o.AllLists)
	for i := 0; i < size-1; i++ {
		for j := i + 1; j < size; j++ {
			equal := true
			for k := 0; k < o.N; k++ {
				if o.AllLists[i][k] != o.AllLists[j][k] {
					equal = false
					break
				}
			}
			if equal {
				duplicated = append(duplicated, j)
			}
		}
	}
	var goodIndex []int
	for i := 0; i < size; i++ {
		good := true
		for _, bad := range duplicated {
			if i == bad {
				good = false
				break
			}
		}
		if good {
			goodIndex = append(goodIndex, i)
		}
	}
	for _, i := range goodIndex {
		o.FinalResult = append(o.FinalResult, o.AllLists[i])
	}
}

func (o *GetAnyN) GetAnyNCards() [][]int {
	var temp []int
	o.ListAllNOfM(o.M, o.N, &temp)
	o.RemoveDuplicated()
	return o.FinalResult
}

// RemoveRepeatedElement 切片去重
func RemoveRepeatedElement(s *[]int) (ret []int) {
	m := make(map[int]int) // key的类型要和切片中的数据类型一致
	var arr []int
	for _, v := range *s {
		m[v] = 1
	}
	for i, _ := range m {
		arr = append(arr, i)
	}
	return arr
}

// IsIncreasedByOne 牌型是否是递增的
func IsIncreasedByOne(seq *[]int) bool {
	lenSize := len(*seq)
	for i := 0; i < lenSize-1; i++ {
		if (*seq)[i+1]-(*seq)[i] != 1 {
			return false
		}
	}
	return true
}

// GetMoveType 返回出牌的类型 例如 一张、一对等
func GetMoveType(move *[]int) []int {
	moveSize := len(*move)
	moveSet := RemoveRepeatedElement(move)
	moveSetSize := len(moveSet)
	if moveSize == 0 {
		//pass 不要
		return []int{TPass, -1}
	}
	if moveSize == 1 {
		//一张
		return []int{TSingle, -1}
	}
	if moveSize == 2 {
		if moveSetSize == 1 {
			//一对
			return []int{TPair, -1}
		} else if moveSetSize == 2 && ((*move)[0] == 20 && (*move)[1] == 30) || ((*move)[0] == 30 && (*move)[1] == 20) {
			//王炸
			return []int{TRocket, -1}
		} else {
			return []int{TWrong, -1}
		}
	}
	if moveSize == 3 {
		//3张一样的
		if moveSetSize == 1 {
			return []int{TTriple, -1}
		} else {
			return []int{TWrong, -1}
		}
	}
	if moveSize == 4 {
		if moveSetSize == 1 {
			//炸弹
			return []int{TBomb, -1}
		} else if moveSetSize == 2 && ((*move)[0] == (*move)[1]) && ((*move)[1] == (*move)[2]) || ((*move)[1] == (*move)[2]) && ((*move)[2] == (*move)[3]) {
			//3带1
			return []int{TTripletSingle, -1}
		} else {
			return []int{TWrong, -1}
		}
	}
	if moveSize == 5 {
		if moveSetSize == 2 {
			//3 带2
			return []int{TTripletPair, -1}
		} else if moveSetSize == 5 && IsIncreasedByOne(move) {
			//顺子
			return []int{TSequenceSingle, 5}
		} else {
			return []int{TWrong, -1}
		}
	}
	moveDict := make(map[int]int)
	for _, card := range *move {
		if _, ok := moveDict[card]; ok {
			moveDict[card]++
		} else {
			moveDict[card] = 1
		}
	}
	countDict := make(map[int]int)
	for _, moveItem := range moveDict {
		if _, ok := countDict[moveItem]; ok {
			countDict[moveItem]++
		} else {
			countDict[moveItem] = 1
		}
	}
	if moveSize == 6 {
		if moveSetSize == 2 {
			if countDict[4] == 1 && countDict[2] == 1 {
				// 4带2
				return []int{TQuadSingle, -1}
			} else if countDict[3] == 1 {
				//连对
				return []int{TSequenceTriple, -1}
			} else {
				return []int{TWrong, -1}
			}
		} else if moveSetSize == 3 {
			if countDict[4] == 1 && countDict[1] == 2 {
				return []int{TQuadSingle, -1}
			} else if countDict[2] == 3 {
				var temp []int
				temp = append(temp, moveSet...)
				sort.Ints(temp)
				if IsIncreasedByOne(&temp) {
					return []int{TSSequencePair, 3} //连对
				}
			} else {
				return []int{TWrong, -1}
			}
		} else if moveSetSize == 6 {
			if countDict[1] == 6 && IsIncreasedByOne(move) {
				return []int{TSequenceSingle, 6} //顺子
			} else {
				return []int{TWrong, -1}
			}
		} else {
			return []int{TWrong, -1}
		}
	}
	if moveSize == 8 && moveSetSize == 3 && countDict[4] == 1 && countDict[2] == 2 {
		return []int{TQuadPair, -1}
	}
	if moveSetSize == countDict[1] && moveSetSize >= MinSingle && IsIncreasedByOne(move) {
		return []int{TSequenceSingle, moveSetSize}
	}
	if moveSetSize == countDict[2] && moveSetSize >= MinPairs {
		var temp []int
		temp = append(temp, moveSet...)
		sort.Ints(temp)
		if IsIncreasedByOne(&temp) {
			return []int{TSSequencePair, moveSetSize}
		}
	}
	if moveSetSize == countDict[3] && moveSetSize >= MinTriples {
		var temp []int
		temp = append(temp, moveSet...)
		sort.Ints(temp)
		if IsIncreasedByOne(&temp) {
			return []int{TSequenceTriple, moveSetSize}
		}
	}
	if countDict[3] >= MinTriples {
		var triples []int
		var singles []int
		var pairs []int
		for key, cardItem := range moveDict {
			if cardItem == 3 {
				triples = append(triples, key)
			} else if cardItem == 2 {
				pairs = append(pairs, key)
			} else if cardItem == 1 {
				singles = append(singles, key)
			} else {
				return []int{TWrong, -1}
			}
		}
		sort.Ints(triples)
		lenTriples := len(triples)
		lenPairs := len(pairs)
		lenSingles := len(singles)
		if IsIncreasedByOne(&triples) {
			if lenTriples == lenSingles && moveSetSize == (lenTriples+lenSingles) {
				return []int{TSequenceTripleSingle, lenTriples}
			}
			if lenTriples == lenPairs && moveSetSize == (lenTriples+lenPairs) {
				return []int{TSequenceTriplePair, lenTriples}
			}
		}
	}
	return []int{TWrong, -1}
}
