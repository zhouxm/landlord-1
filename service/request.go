package service

import (
	"github.com/beego/beego/v2/core/logs"
	"landlord/service/agent"
)

// WSRequest 处理websocket请求
func WSRequest(data []interface{}, client *agent.ClientController) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("WSRequest panic:%v ", r)
		}
	}()
	if len(data) < 1 {
		return
	}
	var req int
	if r, ok := data[0].(float64); ok {
		req = int(r)
	}
	switch req {
	case agent.ReqCheat:
		if len(data) < 2 {
			logs.Error("user [%d] request ReqCheat ,but missing user id", client.User.Id)
			return
		}

	case agent.ReqLogin:
		logs.Debug("ReqLogin %v userid:%v name:%v", agent.RespLogin, client.User.Id, client.User.Username)
		client.SendMsg([]interface{}{agent.RespLogin, client.User.Id, client.User.Username})

	case agent.ReqRoomList:
		logs.Debug("ReqRoomList %v RespRoomList:%v", agent.ReqRoomList, agent.RespRoomList)
		client.SendMsg([]interface{}{agent.RespRoomList})

	case agent.ReqTableList:
		client.SendRoomTables()

	case agent.ReqJoinRoom:
		if len(data) < 2 {
			logs.Error("user [%d] request join room ,but missing room id", client.User.Id)
			return
		}
		var roomId int
		if id, ok := data[1].(float64); ok {
			roomId = int(id)
		}
		roomManager.Lock.RLock()
		defer roomManager.Lock.RUnlock()
		if room, ok := roomManager.Rooms[roomId]; ok {
			client.Room = room
			res := make([][2]int, 0)
			for _, table := range client.Room.Tables {
				if len(table.TableClients) < 3 {
					res = append(res, [2]int{int(table.TableId), len(table.TableClients)})
				}
			}
			logs.Debug("RespJoinRoom %v res:%v", agent.RespJoinRoom, res)
			client.SendMsg([]interface{}{agent.RespJoinRoom, res})
		}

	case agent.ReqNewTable:
		table := client.Room.newTable(client)
		table.joinTable(client)

	case agent.ReqJoinTable:
		if len(data) < 2 {
			return
		}
		var tableId TableId
		if id, ok := data[1].(float64); ok {
			tableId = TableId(id)
		}
		if client.Room == nil {
			return
		}
		client.Room.Lock.RLock()
		defer client.Room.Lock.RUnlock()

		if table, ok := client.Room.Tables[tableId]; ok {
			table.joinTable(client)
		}
		client.SendRoomTables()
	case agent.ReqDealPoker:
		if client.Table.State == GameEnd {
			client.Ready = true
		}
	case agent.ReqCallScore:
		logs.Debug("[%v] ReqCallScore %v", client.User.Username, data)
		client.Table.Lock.Lock()
		defer client.Table.Lock.Unlock()

		if client.Table.State != GameCallScore {
			logs.Debug("game call score at run time ,%v", client.Table.State)
			return
		}
		if client.Table.GameManage.Turn == client || client.Table.GameManage.FirstCallScore == client {
			client.Table.GameManage.Turn = client.Next
		} else {
			logs.Debug("user [%v] call score turn err ")
		}
		if len(data) < 2 {
			return
		}
		var score int
		if s, ok := data[1].(float64); ok {
			score = int(s)
		}

		if score > 0 && score < client.Table.GameManage.MaxCallScore || score > 3 {
			logs.Error("player[%d] call score[%d] cheat", client.User.Id, score)
			return
		}
		if score > client.Table.GameManage.MaxCallScore {
			client.Table.GameManage.MaxCallScore = score
			client.Table.GameManage.MaxCallScoreTurn = client
		}
		client.IsCalled = true
		callEnd := score == 3 || client.Table.allCalled()
		userCall := []interface{}{agent.RespCallScore, client.User.Id, score, callEnd}
		for _, c := range client.Table.TableClients {
			logs.Debug("ReqCallScore response:%v", userCall)
			c.SendMsg(userCall)
		}
		if callEnd {
			logs.Debug("call score end")
			client.Table.callEnd()
		}
	case agent.ReqShotPoker:
		logs.Debug("user [%v] ReqShotPoker %v", client.User.Username, data)
		client.Table.Lock.Lock()
		defer func() {
			client.Table.GameManage.Turn = client.Next
			client.Table.Lock.Unlock()
		}()

		if client.Table.GameManage.Turn != client {
			logs.Error("shot poker err,not your [%d] turn .[%d]", client.User.Id, client.Table.GameManage.Turn.User.Id)
			return
		}
		if len(data) > 1 {
			if pokers, ok := data[1].([]interface{}); ok {
				shotPokers := make([]int, 0, len(pokers))
				for _, item := range pokers {
					if i, ok := item.(float64); ok {
						poker := int(i)
						inHand := false
						for _, handPoker := range client.HandPokers {
							if handPoker == poker {
								inHand = true
								break
							}
						}
						if inHand {
							shotPokers = append(shotPokers, poker)
						} else {
							logs.Warn("player[%d] play non-exist poker", client.User.Id)
							res := []interface{}{agent.RespShotPoker, client.User.Id, []int{}}
							for _, c := range client.Table.TableClients {
								logs.Debug("ReqShotPoker response:%v", res)
								c.SendMsg(res)
							}
							return
						}
					}
				}
				if len(shotPokers) > 0 {
					compareRes, isMulti := ComparePoker(client.Table.GameManage.LastShotPoker, shotPokers)
					if client.Table.GameManage.LastShotClient != client && compareRes < 1 {
						logs.Warn("player[%d] shot poker %v small than last shot poker %v ", client.User.Id, shotPokers, client.Table.GameManage.LastShotPoker)
						res := []interface{}{agent.RespShotPoker, client.User.Id, []int{}}
						for _, c := range client.Table.TableClients {
							logs.Debug("ReqShotPoker response:%v", res)
							c.SendMsg(res)
						}
						return
					}
					if isMulti {
						client.Table.GameManage.Multiple *= 2
					}
					client.Table.GameManage.LastShotClient = client
					client.Table.GameManage.LastShotPoker = shotPokers
					for _, shotPoker := range shotPokers {
						for i, poker := range client.HandPokers {
							if shotPoker == poker {
								copy(client.HandPokers[i:], client.HandPokers[i+1:])
								client.HandPokers = client.HandPokers[:len(client.HandPokers)-1]
								break
							}
						}
					}
				}
				res := []interface{}{agent.RespShotPoker, client.User.Id, shotPokers}
				logs.Debug("ReqShotPoker response:%v", res)
				for _, c := range client.Table.TableClients {
					c.SendMsg(res)
				}
				if len(client.HandPokers) == 0 {
					client.Table.gameOver(client)
				}
			}
		}

		//case ReqGameOver:
	case agent.ReqChat:
		if len(data) > 1 {
			switch data[1].(type) {
			case string:
				client.Table.chat(client, data[1].(string))
			}
		}
	case agent.ReqRestart:
		client.Table.Lock.Lock()
		defer client.Table.Lock.Unlock()

		if client.Table.State == GameEnd {
			client.Ready = true
			for _, c := range client.Table.TableClients {
				if c.Ready == false {
					return
				}
			}
			logs.Debug("restart")
			client.Table.reset()
		}
	}
}
