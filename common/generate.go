package common

import (
	"sort"
)

type void struct{}

var member void

type MovesGenerate struct {
	Cards             []int
	CardsSet          []int //无重复
	CardsDict         map[int]int
	SingleMoves       [][]int
	PairMoves         [][]int
	TripleMoves       [][]int
	BombMoves         [][]int
	KingBombMoves     [][]int
	SerialPairMoves   [][]int
	SerialTripleMoves [][]int
}

func NewMovesGenerate(Cards []int, CardsSet []int, CardsDict map[int]int,
	SingleMoves [][]int, PairMoves [][]int, TripleMoves [][]int, BombMoves [][]int,
	KingBombMoves [][]int, SerialPairMoves [][]int, SerialTripleMoves [][]int) *MovesGenerate {
	return &MovesGenerate{Cards, CardsSet, CardsDict, SingleMoves, PairMoves,
		TripleMoves, BombMoves, KingBombMoves, SerialPairMoves,
		SerialTripleMoves}
}

// Init TODO:需要修改为工厂函数的形式
func (o *MovesGenerate) Init() {
	//o.CardsSet = o.RemoveRepeatedElement(&o.Cards)
	tmpM := make(map[int]void) // key的类型要和切片中的数据类型一致
	for _, v := range o.Cards {
		if _, ok := tmpM[v]; !ok {
			tmpM[v] = member
			o.CardsSet = append(o.CardsSet, v)
		}
	}
	//CardsSet := mapset.NewSetFromSlice([]int{})
	for _, i := range o.Cards {
		if _, ok := o.CardsDict[i]; ok {
			o.CardsDict[i]++
		} else {
			o.CardsDict[i] = 1
		}
	}
}

// RemoveRepeatedElement 这个函数的问题
func (o *MovesGenerate) RemoveRepeatedElement(s *[]int) []int {
	tmpM := make(map[int]void) // key的类型要和切片中的数据类型一致
	var newArr []int
	for _, v := range *s {
		if _, ok := tmpM[v]; !ok {
			tmpM[v] = member
			newArr = append(newArr, v)
		}
	}
	//for i, _ := range tmpM {
	//
	//}
	return newArr
}

func (o *MovesGenerate) IntContains(array []int, val int) (index bool) {
	index = false
	lenArray := len(array)
	for i := 0; i < lenArray; i++ {
		if array[i] == val {
			index = true
			return
		}
	}
	return
}

func (o *MovesGenerate) IntContainsIndex(array []int, val int) (exist bool, index int) {
	exist = false
	index = -1
	lenArray := len(array)
	for i := 0; i < lenArray; i++ {
		if array[i] == val {
			exist = true
			index = i
			return
		}
	}
	return
}

func (o *MovesGenerate) InitCardsDict(cards *[]int) {
	//o.CardsSet = o.RemoveRepeatedElement(&o.Cards) //去重
	tmpM := make(map[int]void) // key的类型要和切片中的数据类型一致
	for _, v := range o.Cards {
		if _, ok := tmpM[v]; !ok {
			tmpM[v] = member
			o.CardsSet = append(o.CardsSet, v)
		}
	}
	for _, i := range *cards {
		if _, ok := o.CardsDict[i]; ok {
			o.CardsDict[i]++
		} else {
			o.CardsDict[i] = 1
		}
	}
}

func (o *MovesGenerate) GenSerialMoves(candidateCards []int, minSerial int, selfType int, serialNum int) [][]int {
	var moves [][]int
	var seqRecords [][]int
	if serialNum < minSerial {
		serialNum = 0
	}
	//candidateCards = o.RemoveRepeatedElement(&candidateCards) //去重
	var tempCandidateCards []int
	tmpM := make(map[int]void) // key的类型要和切片中的数据类型一致
	for _, v := range candidateCards {
		if _, ok := tmpM[v]; !ok {
			tmpM[v] = member
			tempCandidateCards = append(tempCandidateCards, v)
		}
	}
	candidateCards = tempCandidateCards
	sort.Ints(candidateCards) //排序 由小到大
	index := 0
	longest := 1
	start := 0
	sizeSingle := len(candidateCards)
	for index < sizeSingle {
		if (index+1) < sizeSingle && (candidateCards[index+1]-candidateCards[index] == 1) {
			longest++
			index++
		} else {
			// seq_records is like: {{3,2},{4,3},{5,2}}
			seqRecords = append(seqRecords, []int{start, longest})
			index++
			start = index
			longest = 1
		}
	}
	for _, seq := range seqRecords {
		if seq[1] < minSerial { //长度为2
			continue
		}
		start = seq[0]
		longest = seq[1]
		if longest < minSerial {
			continue
		}
		longestList := candidateCards[start : start+longest]
		if serialNum == 0 {
			if len(longestList) < minSerial {
				continue
			}
			for length := minSerial; length <= longest; length++ {
				for startPos := 0; startPos+length <= longest; startPos++ {
					tempMove := longestList[startPos : startPos+length]
					var oneMove []int
					for i := 0; i < selfType; i++ {
						oneMove = append(oneMove, tempMove...) //追加
					}
					sort.Ints(oneMove)
					moves = append(moves, oneMove)
				}
			}
		} else {
			for startPos := 0; startPos+serialNum <= longest; startPos++ {
				tempMove := longestList[startPos : startPos+serialNum]
				var oneMove []int
				for i := 0; i < selfType; i++ {
					oneMove = append(oneMove, tempMove...) //追加
				}
				sort.Ints(oneMove)
				moves = append(moves, oneMove)
			}
		}
	}
	return moves
}

// GenType1Single 1张
func (o *MovesGenerate) GenType1Single() {
	if len(o.SingleMoves) > 0 {
		//return o.SingleMoves
		return
	}
	sort.Ints(o.CardsSet)
	//var temp []int
	//var tempCard = make([]int, 1)
	for _, card := range o.CardsSet {
		temp := []int{card}
		//tempCard[0] = card
		o.SingleMoves = append(o.SingleMoves, temp)
	}
	return
}

// GenType2Pair 一对
func (o *MovesGenerate) GenType2Pair() {
	if len(o.PairMoves) > 0 {
		//return o.PairMoves
		return
	}

	for key, card := range o.CardsDict {
		if card >= 2 {
			temp := []int{key, key}
			o.PairMoves = append(o.PairMoves, temp)
		}
	}
	return
}

// GenType3Triple 3张
func (o *MovesGenerate) GenType3Triple() {
	if len(o.TripleMoves) > 0 {
		//return o.TripleMoves
		return
	}
	for key, card := range o.CardsDict {
		if card >= 3 {
			temp := []int{key, key, key}
			o.TripleMoves = append(o.TripleMoves, temp)
		}
	}
	return
}

// GenType4Bomb 炸弹
func (o *MovesGenerate) GenType4Bomb() {
	if len(o.BombMoves) > 0 {
		return
	}
	//var tempCard = make([]int, 4)
	for key, card := range o.CardsDict {
		if card >= 4 {
			//tempCard[0] = key
			//tempCard[1] = key
			//tempCard[2] = key
			//tempCard[3] = key
			temp := []int{key, key, key, key}
			o.BombMoves = append(o.BombMoves, temp)
		}
	}
	//return o.BombMoves
	return
}

// GenType5KingBomb 王炸
func (o *MovesGenerate) GenType5KingBomb() {
	if len(o.KingBombMoves) > 0 {
		//return o.KingBombMoves
		return
	}
	if o.IntContains(o.CardsSet, 20) && o.IntContains(o.CardsSet, 30) {
		temp := []int{20, 30}
		o.KingBombMoves = append(o.KingBombMoves, temp)
	}
	return
}

// GenType631 3带1
func (o *MovesGenerate) GenType631() [][]int {
	var result [][]int
	if len(o.TripleMoves) == 0 {
		o.GenType3Triple()
	}
	for _, t := range o.TripleMoves {
		for _, c := range o.CardsSet {
			if t[0] != c {
				temp := []int{t[0], t[0], t[0], c}
				result = append(result, temp)
			}
		}
	}
	return result
}

// GenType732 3带2
func (o *MovesGenerate) GenType732() [][]int {
	var result [][]int
	if len(o.TripleMoves) == 0 {
		o.GenType3Triple()
	}
	if len(o.PairMoves) == 0 {
		o.GenType2Pair()
	}
	for _, t := range o.TripleMoves {
		for _, p := range o.PairMoves {
			if t[0] != p[0] {
				temp := []int{t[0], t[0], t[0], p[0], p[0]}
				result = append(result, temp)
			}
		}
	}
	return result
}

// GenType8SerialSingle 8张顺
func (o *MovesGenerate) GenType8SerialSingle(serialNum int) [][]int {
	return o.GenSerialMoves(o.Cards, MinSingle, TypeSingle, serialNum)
}

// GenType9SerialSingle 9张顺
func (o *MovesGenerate) GenType9SerialSingle(serialNum int) [][]int {
	if len(o.SerialPairMoves) > 0 {
		return o.SerialPairMoves
	}
	var singlePairs []int
	for key, card := range o.CardsDict {
		if card >= 2 {
			singlePairs = append(singlePairs, key)
		}
	}
	serialPairMoves := o.GenSerialMoves(singlePairs, MinPairs, TypePair, serialNum)
	return serialPairMoves
}

// GenType10SerialTriple 10张顺
func (o *MovesGenerate) GenType10SerialTriple(serialNum int) [][]int {
	if len(o.SerialTripleMoves) > 0 {
		return o.SerialTripleMoves
	}
	var singleTriple []int
	for key, card := range o.CardsDict {
		if card >= 3 {
			singleTriple = append(singleTriple, key)
		}
	}
	o.SerialTripleMoves = o.GenSerialMoves(singleTriple, MinTriples, TypeTriple, serialNum)
	return o.SerialTripleMoves
}

// GenType11Serial31 11张顺
func (o *MovesGenerate) GenType11Serial31(serialNum int) [][]int {
	var result [][]int
	if len(o.SerialTripleMoves) == 0 {
		return o.GenType10SerialTriple(serialNum)
	}
	for _, t := range o.SerialTripleMoves {
		tempSet := o.RemoveRepeatedElement(&t) //去重

		var restCards []int
		for _, card := range o.Cards {
			if !o.IntContains(tempSet, card) {
				//不包含
				restCards = append(restCards, card)
			}
		}
		gan := NewGetAnyN(restCards, len(restCards), len(tempSet), [][]int{}, [][]int{})
		singleLists := gan.GetAnyNCards()
		for _, singleCards := range singleLists {
			oneResult := t
			oneResult = append(oneResult, singleCards...)
			result = append(result, oneResult)
		}
	}
	return result
}

// GenType12Serial32 3->A
func (o *MovesGenerate) GenType12Serial32(serialNum int) [][]int {
	var result [][]int
	if len(o.SerialTripleMoves) == 0 {
		return o.GenType10SerialTriple(serialNum)
	}
	if len(o.PairMoves) == 0 {
		o.GenType2Pair()
	}
	var pairSet []int
	for _, p := range o.PairMoves {
		pairSet = append(pairSet, p[0])
	}
	pairSet = o.RemoveRepeatedElement(&pairSet) //去重
	for _, s := range o.SerialTripleMoves {
		var tripleSet []int
		tripleSet = append(tripleSet, s...)
		tripleSet = o.RemoveRepeatedElement(&tripleSet)
		tempSet := pairSet
		for _, card := range s {
			exist, index := o.IntContainsIndex(tempSet, card)
			if exist {
				//TODO:	 tempSet删除card
				temp := append(tempSet[:index:index], tempSet[index:]...)
				tempSet = temp
			}
		}
		singlePairCards := tempSet
		gan := NewGetAnyN(singlePairCards, len(tripleSet), len(singlePairCards), [][]int{}, [][]int{})
		singlePairLists := gan.GetAnyNCards()
		for _, singlePair := range singlePairLists {
			oneResult := s
			for i := 0; i < 2; i++ {
				oneResult = append(oneResult, singlePair...)
			}
			result = append(result, oneResult)
		}
	}
	return result
}

func (o *MovesGenerate) GenType1342() [][]int {
	var result [][]int
	if len(o.BombMoves) == 0 {
		o.GenType4Bomb()
	}
	var allTwoCards [][]int
	lenCards := len(o.Cards)
	if lenCards >= 1 {
		for i := 0; i < lenCards-1; i++ {
			for j := i + 1; j < lenCards; j++ {
				var temp = []int{o.Cards[i], o.Cards[j]}
				allTwoCards = append(allTwoCards, temp)
			}
		}
	}
	var duplicated []int //集合  无重复
	lenAllTwoCards := len(allTwoCards)
	if lenAllTwoCards >= 1 {
		for i := 0; i < lenAllTwoCards-1; i++ {
			for j := i + 1; j < lenAllTwoCards; j++ {
				if allTwoCards[i][0] == allTwoCards[j][0] && allTwoCards[i][1] == allTwoCards[j][1] {
					duplicated = append(duplicated, j)
				}
			}
		}
	}
	duplicated = o.RemoveRepeatedElement(&duplicated)
	var refinedTwoCards [][]int
	for i := 0; i < lenAllTwoCards; i++ {
		if !o.IntContains(duplicated, i) {
			refinedTwoCards = append(refinedTwoCards, allTwoCards[i])
		}
	}
	for _, bomb := range o.BombMoves {
		for _, twoCards := range refinedTwoCards {
			if bomb[0] != twoCards[0] && bomb[0] != twoCards[1] {
				oneResult := bomb
				oneResult = append(oneResult, twoCards...)
				result = append(result, oneResult)
			}
		}
	}
	return result
}

func (o *MovesGenerate) GenType14422() [][]int {
	var result [][]int
	if len(o.BombMoves) == 0 {
		o.GenType4Bomb()
	}
	var twoMoreCards []int
	for key, card := range o.CardsDict {
		if card == 2 || card == 3 {
			twoMoreCards = append(twoMoreCards, key)
		}
	}
	if len(twoMoreCards) < 2 {
		return result
	}
	gan := NewGetAnyN(twoMoreCards, 2, len(twoMoreCards), [][]int{}, [][]int{})
	twoPairs := gan.GetAnyNCards()
	for _, bomb := range o.BombMoves {
		for _, twoPair := range twoPairs {
			if bomb[0] != twoPair[0] && bomb[0] != twoPair[1] {
				oneResult := bomb
				var temp = []int{twoPair[0], twoPair[0], twoPair[1], twoPair[1]}
				oneResult = append(oneResult, temp...)
				result = append(result, oneResult)
			}
		}
	}
	return result
}

// GenAllMoves 得到所有可能的出牌
func (o *MovesGenerate) GenAllMoves() MySortInt {
	var result MySortInt
	var temp [][]int
	o.GenType1Single()
	result = append(result, o.SingleMoves...)
	o.GenType2Pair()
	result = append(result, o.PairMoves...)
	o.GenType3Triple()
	result = append(result, o.TripleMoves...)
	o.GenType4Bomb()
	result = append(result, o.BombMoves...)
	o.GenType5KingBomb()
	result = append(result, o.KingBombMoves...)

	temp = o.GenType631()
	result = append(result, temp...)
	temp = o.GenType732()
	result = append(result, temp...)
	temp = o.GenType8SerialSingle(0)
	result = append(result, temp...)
	temp = o.GenType9SerialSingle(0)
	result = append(result, temp...)
	temp = o.GenType10SerialTriple(0)
	result = append(result, temp...)
	temp = o.GenType11Serial31(0)
	result = append(result, temp...)
	temp = o.GenType12Serial32(0)
	result = append(result, temp...)
	temp = o.GenType1342()
	result = append(result, temp...)
	temp = o.GenType14422()
	result = append(result, temp...)
	return result
}
