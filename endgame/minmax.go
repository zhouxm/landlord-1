package endgame

import (
	"errors"
	"github.com/beego/beego/v2/core/logs"
	"landlord/common"
	"sort"
	"sync"
	"time"
)

const (
	PEASANT  = 0
	LANDLORD = 1
	MinScore = 0
	MaxScore = 999
)

var rfPool = &sync.Pool{New: func() interface{} {
	return common.NewMovesGenerate([]int{}, []int{}, make(map[int]int), [][]int{},
		[][]int{}, [][]int{}, [][]int{},
		[][]int{}, [][]int{}, [][]int{})
}}

type BestResult struct {
	Score    int
	BestMove []int
}

func GetProperMoves(cards *[]int, rivalMove *[]int) (common.MySortInt, error) {
	var properMoves common.MySortInt
	var allMoves [][]int
	moveTypePair := common.GetMoveType(rivalMove)
	moveType := moveTypePair[0]
	serialNum := moveTypePair[1]
	//TODO :测试这里
	//mg := NewMovesGenerate(*cards, []int{}, make(map[int]int), [][]int{},
	//	[][]int{}, [][]int{}, [][]int{},
	//	[][]int{}, [][]int{}, [][]int{})
	mg := rfPool.Get().(*common.MovesGenerate) //类型断言
	mg.Cards = *cards
	mg.Init()
	//////
	if moveType == common.TPass {
		properMoves = mg.GenAllMoves()
	} else if moveType == common.TSingle {
		mg.GenType1Single()
		properMoves = FilterType1Single(mg.SingleMoves, rivalMove)
	} else if moveType == common.TPair {
		mg.GenType2Pair()
		properMoves = FilterType2Pair(mg.PairMoves, rivalMove)
	} else if moveType == common.TTriple {
		mg.GenType3Triple()
		properMoves = FilterType3Triple(mg.TripleMoves, rivalMove)
	} else if moveType == common.TBomb {
		mg.GenType4Bomb()
		mg.GenType5KingBomb()
		allMoves = append(allMoves, mg.KingBombMoves...)
		allMoves = append(allMoves, mg.BombMoves...)
		properMoves = FilterType4Bomb(allMoves, rivalMove)
	} else if moveType == common.TRocket {
		properMoves = make([][]int, 1)
	} else if moveType == common.TTripletSingle {
		allMoves = mg.GenType631()
		properMoves = FilterType631(allMoves, rivalMove)
	} else if moveType == common.TTripletPair {
		allMoves = mg.GenType732()
		properMoves = FilterType732(allMoves, rivalMove)
	} else if moveType == common.TSequenceSingle {
		allMoves = mg.GenType8SerialSingle(serialNum)
		properMoves = FilterType8SerialSingle(allMoves, rivalMove)
	} else if moveType == common.TSSequencePair {
		allMoves = mg.GenType9SerialSingle(serialNum)
		properMoves = FilterType9SerialPair(allMoves, rivalMove)
	} else if moveType == common.TSequenceTriple {
		allMoves = mg.GenType10SerialTriple(serialNum)
		properMoves = FilterType10SerialTriple(allMoves, rivalMove)
	} else if moveType == common.TSequenceTripleSingle {
		allMoves = mg.GenType11Serial31(serialNum)
		properMoves = FilterType11Serial31(allMoves, rivalMove)
	} else if moveType == common.TSequenceTriplePair {
		allMoves = mg.GenType12Serial32(serialNum)
		properMoves = FilterType12Serial32(allMoves, rivalMove)
	} else if moveType == common.TQuadSingle {
		allMoves = mg.GenType1342()
		properMoves = FilterType1342(allMoves, rivalMove)
	} else if moveType == common.TQuadPair {
		allMoves = mg.GenType14422()
		properMoves = FilterType14422(allMoves, rivalMove)
	} else {
		logs.Error("invalid endgame type, ", moveType)
		errMsg := errors.New("invalid endgame type ")
		return properMoves, errMsg
	}
	if moveType != common.TPass && moveType != common.TBomb && moveType != common.TRocket {
		mg.GenType4Bomb()
		mg.GenType5KingBomb()
		properMoves = append(properMoves, mg.KingBombMoves...)
		properMoves = append(properMoves, mg.BombMoves...)
	}
	if len(*rivalMove) != 0 {
		properMoves = append(properMoves, []int{})
	}
	return properMoves, nil
}

type MinMaxObject struct {
	QuitChannel     chan struct{}
	BestMoveChannel chan []int
}

func NewMinMaxObject(QuitChannel chan struct{}, BestMoveChannel chan []int) *MinMaxObject {
	return &MinMaxObject{QuitChannel, BestMoveChannel}
}

// StartEngine channel若读取不到 会进入waiting状态 直到有可读取的值时才会唤醒
// 通道被关闭是可以检测到的
func (o *MinMaxObject) StartEngine(farmerCards []int, lorderCards []int, lastMove []int, currentPlayer int) (*BestResult, error) {
	defer func() {
		err := recover()
		if err != nil {
			logs.Error("[Except] StartEngine err :", err)
		}
		close(o.BestMoveChannel) //最后关闭channel
	}()

	mainStart := time.Now()
	properMove, err := GetProperMoves(&lorderCards, &lastMove)
	sort.Sort(properMove)
	logs.Info("StartEngine properMove", properMove)
	if err != nil {
		return &BestResult{MaxScore, []int{}}, err
	}
	moveNum := 0
	for _, indexMove := range properMove {
		restLorderCards := common.GetRestCards(&lorderCards, &indexMove)
		logs.Info("goroutine: ", moveNum)
		logs.Info("restLorderCards: ", restLorderCards, "endgame: ", indexMove)
		moveNum++
		go func(farmerCards []int, restLorderCards []int, lastMove []int, currentPlayer int, o *MinMaxObject) {
			//当一个协程寻找到最佳解后就结束其他的协程 需不需要限制并发的数目？暂时先不限制先试试 2020-02-18
			//找到一个最游记后通知其他的协程退出
			gStartTime := time.Now()

			//检查一下是否结束 g
			if o.Cancels() {
				return
			}

			var nodeCount int64 = 0
			score := o.MinMaxSearch(farmerCards, restLorderCards, lastMove, currentPlayer, &nodeCount)
			// 结束 g 此处是必须
			if o.CheckStatus(nodeCount, true) {
				return
			}

			//打印这个move以及其评分
			logs.Info("score = ", score, "calculated move is: ", lastMove)
			if score == MaxScore {
				o.BestMoveChannel <- lastMove
				logs.Info("Best move is ", lastMove)
				//发送端关闭通道 尽量不在接收端关闭，即使关闭了，接收端也可以接收 多次关闭一个通道会有问题
				//不要在多个并行的生产者端关闭channel ，这里就有这个问题问题，所以需要找合适的方法
			}
			gEndTime := time.Since(gStartTime)
			logs.Info("endgame", lastMove, "time const: ", gEndTime)
			return
		}(farmerCards, restLorderCards, indexMove, PEASANT, o)
	}

	mainEndTime := time.Since(mainStart)
	logs.Info("Main time const: ", mainEndTime)
	if bestMove, ok := <-o.BestMoveChannel; ok {
		close(o.QuitChannel) //重复关闭的问题 只有一个消费者可在消费端关闭
		//处理最优解
		findBestTime := time.Since(mainStart) //毫秒
		logs.Info("-------------------------------")
		logs.Info("StartEngine find best success! Time const: ", findBestTime)
		return &BestResult{MaxScore, bestMove}, nil
	} else {
		// 最优解通道被意外关闭
		logs.Error("StartEngine find best fail!")
		return &BestResult{MinScore, []int{}}, nil
	}
}

func (o *MinMaxObject) SingleThread(farmerCards []int, restLorderCards []int, lastMove []int, currentPlayer int) (int, []int) {
	var nodeCount int64
	nodeCount = 0
	score := o.MinMaxSearch(farmerCards, restLorderCards, lastMove, currentPlayer, &nodeCount)
	//打印这个move以及其评分
	logs.Info("score = ", score, "calculated endgame is: ", lastMove)
	if score == MaxScore {
		return score, lastMove
	}
	return MinScore, []int{}
}

func (o *MinMaxObject) Cancels() bool {
	select {
	case <-o.QuitChannel:
		return true
	default:
		return false
	}
}

func (o *MinMaxObject) MinMaxSearch(farmerCards []int, lorderCards []int, lastMove []int, currentPlayer int, nodeCount *int64) int {
	defer func() {
		//恢复程序的控制权
		err := recover()
		if err != nil {
			logs.Error("[Exception] MinMaxSearch err : ", err)
		}
	}()
	//当前出牌的是地主
	if currentPlayer == LANDLORD && len(farmerCards) == 0 {
		*nodeCount++
		if o.CheckStatus(*nodeCount, false) {
			//logs.Info("LANDLORD quit")
			//此时是已经找到了一个最优解，此处就应该直接 quit goroutine
			//return MaxScore
			return MinScore
		}
		return MinScore
	}
	//当前出牌的是农民
	if currentPlayer == PEASANT && len(lorderCards) == 0 {
		*nodeCount++
		if o.CheckStatus(*nodeCount, false) {
			//logs.Info("PEASANT quit")
			return MaxScore
			//return MinScore
		}
		return MaxScore
	}

	if currentPlayer == LANDLORD { // currentPlayer == LANDLORD 地主
		properMoves, _ := GetProperMoves(&lorderCards, &lastMove)
		sort.Sort(properMoves)
		for _, indexMove := range properMoves {
			lorderCurrentCards := common.GetRestCards(&lorderCards, &indexMove)
			score := o.MinMaxSearch(farmerCards, lorderCurrentCards, indexMove, PEASANT, nodeCount)
			if o.CheckStatus(*nodeCount, false) {
				//logs.Info("currentPlayer LANDLORD quit")
				break
			}
			if score == MaxScore {
				return MaxScore
			}
		}
		return MinScore
	} else { // currentPlayer == PEASANT 农民
		properMoves, _ := GetProperMoves(&farmerCards, &lastMove)
		sort.Sort(properMoves)
		for _, indexMove := range properMoves {
			farmerCurrentCards := common.GetRestCards(&farmerCards, &indexMove)
			score := o.MinMaxSearch(farmerCurrentCards, lorderCards, indexMove, LANDLORD, nodeCount)
			if o.CheckStatus(*nodeCount, false) {
				//此时是已经找到了一个最优解，此处就应该直接结束
				//logs.Info("currentPlayer PEASANT quit")
				break
			}
			if score == MinScore {
				return MinScore
			}
		}
		return MaxScore
	}
}

func (o *MinMaxObject) CheckStatus(nodeCode int64, lastCheck bool) bool {
	if lastCheck {
		logs.Info(nodeCode, " nodes have been calculated...")
	} else {
		if nodeCode%30000 == 0 {
			//logs.Info(nodeCode, " nodes have been calculated...")
		}
	}
	if o.Cancels() {
		return true
	} else {
		return false
	}
}
