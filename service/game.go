package service

type GameManage struct {
	Turn             *ClientController
	FirstCallScore   *ClientController //每局轮转
	MaxCallScore     int               //最大叫分
	MaxCallScoreTurn *ClientController
	LastShotClient   *ClientController
	Pokers           []int
	LastShotPoker    []int
	Multiple         int //加倍
}

type Game struct {
	Turn             *ClientController
	FirstCallScore   *ClientController //每局轮转
	MaxCallScore     int               //最大叫分
	MaxCallScoreTurn *ClientController
	LastShotClient   *ClientController
	players=
}

//TableClients map[int]*ClientController
func (game *Game) AllCalled() bool {
	for _, client := range game.players {
		if !client.IsCalled {
			return false
		}
	}
	return true
}

//// 一局结束
//func (table *Game) GameOver(client *ClientController) {
//	coin := table.Creator.Room.EntranceFee * table.GameManage.MaxCallScore * table.GameManage.Multiple
//	table.State = GameEnd
//	for _, c := range table.TableClients {
//		res := []interface{}{RespGameOver, client.User.Id}
//		if client == c {
//			res = append(res, coin*2-100)
//		} else {
//			res = append(res, coin)
//		}
//		for _, cc := range table.TableClients {
//			if cc != c {
//				userPokers := make([]int, 0, len(cc.HandPokers)+1)
//				userPokers = append(append(userPokers, cc.User.Id), cc.HandPokers...)
//				res = append(res, userPokers)
//			}
//		}
//		logs.Debug("gameOver SendMsg:%v", res)
//		c.SendMsg(res)
//	}
//	logs.Debug("table[%d] game over", table.TableId)
//}
//
//// 叫分阶段结束
//func (table *Game) CallEnd() {
//	table.State = GamePlaying
//	table.GameManage.FirstCallScore = table.GameManage.FirstCallScore.Next
//	if table.GameManage.MaxCallScoreTurn == nil || table.GameManage.MaxCallScore == 0 {
//		table.GameManage.MaxCallScoreTurn = table.Creator
//		table.GameManage.MaxCallScore = 1
//		//return
//	}
//	landLord := table.GameManage.MaxCallScoreTurn
//	landLord.User.Role = RoleLandlord
//	table.GameManage.Turn = landLord
//	for _, poker := range table.GameManage.Pokers {
//		landLord.HandPokers = append(landLord.HandPokers, poker)
//	}
//	res := []interface{}{RespShowPoker, landLord.User.Id, table.GameManage.Pokers}
//	for _, c := range table.TableClients {
//		logs.Debug("callEnd SendMsg:%v", res)
//		c.SendMsg(res)
//	}
//}
