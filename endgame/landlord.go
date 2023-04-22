package endgame

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"landlord/common"
	"landlord/models"
	"sort"
)

// Move 获取出牌
func Move(reqMove *models.ReqMove) (models.RspMove, error) {
	// 必须先检查农民的出牌是否正确，然后才可以根据农民的出牌得到地主的出牌
	logs.Info("reqMove.FirstMove ", reqMove.FirstMove)
	logs.Info("地主家的牌: ", reqMove.LandlordHands)
	logs.Info("农民家的牌: ", reqMove.PeasantHands)
	logs.Info("------------------------------")
	landlordHands, err := common.FormatInputCards(reqMove.LandlordHands) //字符型映射为数值型
	if err != nil {
		return *models.NewRespMove(
			[]string{},
			[]string{},
			[]string{},
			[]string{},
			err.Error(), 400), err
	}
	peasantHands, err := common.FormatInputCards(reqMove.PeasantHands)
	if err != nil {
		return *models.NewRespMove(
			[]string{},
			[]string{},
			[]string{},
			[]string{},
			err.Error(), 400), err
	}
	var result models.RspMove
	if reqMove.FirstMove {
		//第一次出牌 地主先出
		var lastMove []int
		minMaxObj := NewMinMaxObject(make(chan struct{}), make(chan []int, 1))
		resultMove, _ := minMaxObj.StartEngine(peasantHands, landlordHands, lastMove, LANDLORD)
		if resultMove.Score == MinScore {
			logs.Info("地主必败!!")
			return *models.NewRespMove(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					"地主必败", 200),
				fmt.Errorf("地主必败")
		}
		var lorderMove []int
		lorderMove = resultMove.BestMove
		landlordHands = common.GetRestCards(&landlordHands, &lorderMove) //剔除出的牌之后地主剩余的牌
		result.LandlordMove = common.FormatOutputCards(lorderMove)       //地主出的牌
		result.PeasantHands = common.FormatOutputCards(peasantHands)     //第一把出牌时农民出的牌为空
		result.LandlordHands = common.FormatOutputCards(landlordHands)   //地主底牌牌
		result.PeasantMove = []string{}
		result.Msg = ""
		result.Code = 200
		return result, nil
	} else {
		//不是第一次出牌
		var peasantMove []int
		var err error
		//得到农民出的牌
		if len(reqMove.PeasantMove) == 0 || reqMove.PeasantMove[0] == "pass" || reqMove.PeasantMove[0] == "Pass" {
			peasantMove = []int{}
		} else {
			peasantMove, err = common.FormatInputCards(reqMove.PeasantMove)
			if err != nil {
				return *models.NewRespMove(
					reqMove.LandlordHands,
					reqMove.PeasantHands,
					reqMove.LandlordMove,
					reqMove.PeasantMove,
					err.Error(), 400), err
			}
		}
		//根据地主出的牌和农民的手牌找到所有可能的出牌，验证农民的出牌是否正确
		landlordMove, _ := common.FormatInputCards(reqMove.LandlordMove) //地主上一手出的牌
		properMoves, _ := GetProperMoves(&peasantHands, &landlordMove)   //根据地主的出牌和农民的底牌得到所有的可能出牌

		//校验农民的出牌是否存在于properMoves
		logs.Info("proper properMoves: ", properMoves)
		farmerMoveExists := check(peasantMove, properMoves)

		if !farmerMoveExists {
			//若不存在直接返回
			return *models.NewRespMove(
				reqMove.LandlordHands,
				reqMove.PeasantHands,
				reqMove.LandlordMove,
				reqMove.PeasantMove,
				"出牌错误!!请重新出牌", 400), fmt.Errorf("出牌错误!!请重新出牌")
		}

		peasantHands = common.GetRestCards(&peasantHands, &peasantMove) //得到农民该轮出牌后剩余的牌
		if len(peasantHands) == 0 {
			// 农民牌出完手里的牌就出完了
			return *models.NewRespMove(
				reqMove.LandlordHands, //地主就没必要出了
				[]string{},
				reqMove.LandlordMove,
				common.FormatOutputCards(peasantMove),
				"农民胜利!!", 200), nil
		}

		if len(peasantMove) > 0 {
			result.PeasantMove = common.FormatOutputCards(peasantMove)
		} else {
			result.PeasantMove = []string{"Pass"} //农民不出
		}

		result.PeasantHands = common.FormatOutputCards(peasantHands) //农民出牌后的底牌
		minMaxObj := NewMinMaxObject(make(chan struct{}), make(chan []int, 1))
		resultMove, _ := minMaxObj.StartEngine(peasantHands, landlordHands, peasantMove, LANDLORD)
		if resultMove.Score == MinScore {
			return *models.NewRespMove(
				reqMove.LandlordHands,
				reqMove.PeasantHands,
				reqMove.LandlordMove,
				reqMove.PeasantMove,
				"地主必败，不用玩了!!", 200), nil
		}
		landlordHands = common.GetRestCards(&landlordHands, &resultMove.BestMove) //地主出了牌之后剩的牌
		result.LandlordMove = common.FormatOutputCards(resultMove.BestMove)
		if len(landlordHands) == 0 {
			//地主出完这轮牌就赢了
			result.LandlordHands = []string{}
			result.Msg = "地主胜利!!"
			result.Code = 200
		} else {
			result.LandlordHands = common.FormatOutputCards(landlordHands)
			result.Code = 200
		}
		return result, nil
	}
}

// 验证出牌是否正确
func check(divMove []int, properMoves [][]int) bool {
	if len(divMove) == 0 {
		return true
	} else {
		sort.Ints(divMove)
		for _, move := range properMoves {
			if len(move) != len(divMove) {
				continue
			}
			sort.Ints(move)
			allSame := true
			for i := 0; i < len(move); i++ {
				if divMove[i] != move[i] {
					allSame = false
					break
				}
			}
			if allSame == true {
				return true
			}
		}
	}
	return false
}
