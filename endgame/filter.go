package endgame

import (
	"landlord/common"
	"sort"
)

func CommonHandle(moves [][]int, rivalMove *[]int) common.MySortInt {
	var newMoves common.MySortInt
	for _, move := range moves {
		if move[0] > (*rivalMove)[0] {
			newMoves = append(newMoves, move)
		}
	}
	return newMoves
}

func FilterType1Single(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType2Pair(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType3Triple(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType4Bomb(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType631(moves [][]int, rivalMove *[]int) [][]int {
	var filteredMoves [][]int
	var targetRivalCard = 999
	rivalDict := make(map[int]int)
	for _, card := range *rivalMove {
		if _, ok := rivalDict[card]; ok {
			targetRivalCard = card
			break
		} else {
			rivalDict[card] = 1
		}
	}
	for _, move := range moves {
		moveDict := make(map[int]int)
		for _, card := range move {
			if _, ok := moveDict[card]; ok {
				if card > targetRivalCard {
					filteredMoves = append(filteredMoves, move)
				}
				break
			} else {
				moveDict[card] = 1
			}
		}
	}
	return filteredMoves
}

func FilterType732(moves [][]int, rivalMove *[]int) [][]int {
	var filteredMoves [][]int
	var targetRivalCard = 999
	rivalDict := make(map[int]int)
	for _, card := range *rivalMove {
		if _, ok := rivalDict[card]; ok {
			rivalDict[card]++
			if rivalDict[card] == 3 {
				targetRivalCard = card
				break
			}
		} else {
			rivalDict[card] = 1
		}
	}
	for _, move := range moves {
		moveDict := make(map[int]int)
		for _, card := range move {
			if _, ok := moveDict[card]; ok {
				moveDict[card]++
				if moveDict[card] == 3 {
					if card > targetRivalCard {
						filteredMoves = append(filteredMoves, move)
					}
					break
				}
			} else {
				//未找到
				moveDict[card] = 1
			}
		}
	}
	return filteredMoves
}

func FilterType8SerialSingle(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType9SerialPair(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType10SerialTriple(moves [][]int, rivalMove *[]int) [][]int {
	return CommonHandle(moves, rivalMove)
}

func FilterType11AndType12(moves [][]int, rivalMove *[]int) [][]int {
	var filteredMoves [][]int
	rivalMoveDict := make(map[int]int)
	var rivalTriples []int
	for _, card := range *rivalMove {
		if _, ok := rivalMoveDict[card]; ok {
			rivalMoveDict[card]++
			if rivalMoveDict[card] == 3 {
				rivalTriples = append(rivalTriples, card)
			}
		} else {
			rivalMoveDict[card] = 1
		}
	}
	sort.Ints(rivalTriples)
	for _, move := range moves {
		var moveTriples []int
		moveDict := make(map[int]int)
		for _, card := range move {
			if _, ok := moveDict[card]; ok {
				moveDict[card]++
				if moveDict[card] == 3 {
					moveTriples = append(moveTriples, card)
				}
			} else {
				//未找到
				moveDict[card] = 1
			}
		}
		sort.Ints(moveTriples)
		if moveTriples[0] > rivalTriples[0] {
			filteredMoves = append(filteredMoves, move)
		}
	}
	return filteredMoves
}

func FilterType11Serial31(moves [][]int, rivalMove *[]int) [][]int {
	return FilterType11AndType12(moves, rivalMove)
}

func FilterType12Serial32(moves [][]int, rivalMove *[]int) [][]int {
	return FilterType11AndType12(moves, rivalMove)
}

func FilterType13AndType14(moves [][]int, rivalMove *[]int) [][]int {
	var filteredMoves [][]int
	rivalDict := make(map[int]int)
	targetRivalCard := 999
	for _, card := range *rivalMove {
		if _, ok := rivalDict[card]; ok {
			rivalDict[card]++
			if rivalDict[card] == 4 {
				targetRivalCard = card
				break
			}
		} else {
			rivalDict[card] = 1
		}
	}
	for _, move := range moves {
		bombCard := -1
		moveDict := make(map[int]int)
		for _, card := range move {
			if _, ok := moveDict[card]; ok {
				moveDict[card]++
				if moveDict[card] == 4 {
					bombCard = card
					break
				}
			} else {
				//未找到
				moveDict[card] = 1
			}
		}
		if bombCard > targetRivalCard {
			filteredMoves = append(filteredMoves, move)
		}
	}
	return filteredMoves
}

func FilterType1342(moves [][]int, rivalMove *[]int) [][]int {
	return FilterType13AndType14(moves, rivalMove)
}

func FilterType14422(moves [][]int, rivalMove *[]int) [][]int {
	return FilterType13AndType14(moves, rivalMove)
}
