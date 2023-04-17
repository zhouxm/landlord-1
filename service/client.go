package service

import (
	"bytes"
	"encoding/json"
	"landlord/models"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"

	"github.com/beego/beego/v2/core/logs"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	RoleFarmer   = 0
	RoleLandlord = 1
)

var (
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	} //不验证origin
)

type ClientController struct {
	web.Controller
	conn       *websocket.Conn
	User       *models.Account
	Room       *Room
	Table      *Table
	HandPokers []int
	Ready      bool
	IsCalled   bool              //是否叫完分
	Next       *ClientController //链表
	IsRobot    bool
	ToRobot    chan []interface{} //发送给robot的消息
	ToServer   chan []interface{} //robot发送给服务器
}

// Reset 重置状态
func (c *ClientController) Reset() {
	c.User.Role = 1
	c.HandPokers = make([]int, 0, 21)
	c.Ready = false
	c.IsCalled = false
}

// RspRoomTables 将房间内已有的牌桌信息，发送给客户端
func (c *ClientController) RspRoomTables() {
	res := make([][2]int, 0)
	for _, table := range c.Room.Tables {
		if len(table.Clients) < 3 {
			res = append(res, [2]int{int(table.TableId), len(table.Clients)})
		}
	}
	c.SendMessageToClient([]interface{}{RspTableList, res})
}

// SendMessageToClient Send message to client，if the client is robot，push the message to the 'ToRobot' channel, else use zhe  websocket conn write
func (c *ClientController) SendMessageToClient(message []interface{}) {

	if c.IsRobot {
		c.ToRobot <- message
		return
	}
	msgByte, err := json.Marshal(message)
	if err != nil {
		logs.Error("send msg [%v] marsha1 err:%v", string(msgByte), err)
		return
	}
	err = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		logs.Error("send msg SetWriteDeadline [%v] err:%v", string(msgByte), err)
		return
	}
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		err = c.conn.Close()
		if err != nil {
			logs.Error("close client err: %v", err)
		}
	}
	_, err = w.Write(msgByte)
	if err != nil {
		logs.Error("Write msg [%v] err: %v", string(msgByte), err)
	}
	if err := w.Close(); err != nil {
		err = c.conn.Close()
		if err != nil {
			logs.Error("close err: %v", err)
		}
	}
}

// 关闭客户端
// func (c *ClientController) close() {
// 	if c.Table != nil {
// 		for _, client := range c.Table.TableClients {
// 			if c.Table.Creator == c && c != client {
// 				c.Table.Creator = client
// 			}
// 			if c == client.Next {
// 				client.Next = nil
// 			}
// 		}
// 		if len(c.Table.TableClients) != 1 {
// 			for _, client := range c.Table.TableClients {
// 				if client != client.Table.Creator {
// 					client.Table.Creator.Next = client
// 				}
// 			}
// 		}
// 		if len(c.Table.TableClients) == 1 {
// 			c.Table.Creator = nil
// 			delete(c.Room.Tables, c.Table.TableId)
// 			return
// 		}
// 		delete(c.Table.TableClients, c.User.Id)
// 		if c.Table.State == GamePlaying {
// 			c.Table.SyncUser()
// 			//c.Table.Reset()
// 		}
// 		if c.IsRobot {
// 			close(c.ToRobot)
// 			close(c.ToServer)
// 		}
// 	}
// }

// Ping 心跳
func (c *ClientController) Ping() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logs.Warn(err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logs.Warn(err)
			}
		}
	}
}

func (c *ClientController) ServeWs() {
	conn, err := upGrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		logs.Error("upGrader err:%v", err)
		return
	}
	client := &ClientController{conn: conn, HandPokers: make([]int, 0, 21), User: &models.Account{}}
	if userId, err := strconv.Atoi(c.Ctx.GetCookie("userid")); err != nil {
		logs.Error(err)
	} else {
		client.User.Id = userId
	}

	username := c.Ctx.GetCookie("username")
	if username != "" {
		client.User.Username = username
	}

	go client.RecvWebSocketMessage()
	go client.Ping()
}

func (c *ClientController) RecvWebSocketMessage() {

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Error("websocket user_id[%d] unexpected close error: %v", c.User.Id, err)
			}
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var data []interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			logs.Error("message Unmarshal err, user_id[%d] err:%v", c.User.Id, err)
		} else {
			WSRequest(data, c)
		}
	}
}

// RecvRobotChannelMessage Robot Message Receive
func (c *ClientController) RecvRobotChannelMessage() {
	for {
		select {
		case message, ok := <-c.ToServer:
			if !ok {
				return
			}
			WSRequest(message, c)
		case message, ok := <-c.ToRobot:
			if !ok {
				return
			}
			logs.Debug("RecvRobotChannelMessage Player [%v] Receive Message %v ", c.User.Username, message)
			if len(message) < 1 {
				logs.Error("RecvRobotChannelMessage Player [%v] Receive Message Error, Message: %v", c.User.Username, message)
				return
			}
			if act, ok := message[0].(int); ok {
				//protocolCode := act
				switch act {
				case RspDealPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.FirstCallScore == c {
						c.autoCallScore()
					}
					c.Table.Lock.RUnlock()

				case RspCallScore:
					if len(message) < 4 {
						logs.Error("RecvRobotChannelMessage:RepCallScore msg err:%v", message)
						return
					}
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c && !c.IsCalled {
						var callEnd bool
						logs.Debug("RecvRobotChannelMessage:RepCallScore %t", message)
						if res, ok := message[3].(bool); ok {
							callEnd = res
						}
						if !callEnd {
							c.autoCallScore()
						}
					}
					c.Table.Lock.RUnlock()

				case RspShotPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()

				case RspShowPoker:
					time.Sleep(time.Second)
					//logs.Debug("robot [%v] role [%v] receive message RepShowPoker turn :%v", c.User.Username, c.User.Role, c.Table.GameManage.Turn.User.Username)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c || (c.Table.GameManage.Turn == nil && c.User.Role == RoleLandlord) {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()
				case RspGameOver:
					c.Ready = true
				}
			} else {
				logs.Error("client.ToRobot msg[0] is not int, the message:%v", message)
			}

		}
	}
}

// 自动出牌
func (c *ClientController) autoShotPoker() {
	//因为机器人休眠一秒后才出牌，有可能因用户退出而关闭chan
	defer func() {
		err := recover()
		if err != nil {
			logs.Warn("autoShotPoker err : %v", err)
		}
	}()
	shotPokers := make([]int, 0)
	if len(c.Table.GameManage.LastShotPoker) == 0 || c.Table.GameManage.LastShotClient == c {
		shotPokers = append(shotPokers, c.HandPokers[0])
	} else {
		shotPokers = CardsAbove(c.HandPokers, c.Table.GameManage.LastShotPoker)
	}
	pokers := make([]interface{}, 0)
	for _, poker := range shotPokers {
		pokers = append(pokers, poker)
	}
	message := []interface{}{float64(ReqShotPoker)}
	message = append(message, pokers)
	logs.Debug("robotID:%v autoShotPoker :%v", c.User.Id, pokers)
	c.ToServer <- message
}

// 自动叫分
func (c *ClientController) autoCallScore() {
	defer func() {
		err := recover()
		if err != nil {
			logs.Warn("autoCallScore err : %v", err)
		}
	}()
	logs.Debug("robot [%v] autoCallScore", c.User.Username)
	c.ToServer <- []interface{}{float64(ReqCallScore), int(3)}
}

// func rece

// WSRequest 处理websocket请求
func WSRequest(data []interface{}, client *ClientController) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("WSRequest panic:%v ", r)
		}
	}()
	if len(data) < 1 {
		return
	}
	req, ok := data[0].(float64)
	if !ok {
		logs.Error("WSRequest data[0] to float64 error, the data:%T", data[0])
		return
	}
	switch req {
	case ReqCheat:
		if len(data) < 2 {
			logs.Error("WSRequest user [%d] request ReqCheat ,but missing user id", client.User.Id)
			return
		}

	case ReqLogin:
		logs.Debug("ReqLogin %v userid:%v name:%v", RspLogin, client.User.Id, client.User.Username)
		client.SendMessageToClient([]interface{}{RspLogin, client.User.Id, client.User.Username})

	case ReqRoomList:
		logs.Debug("ReqRoomList %v RepRoomList:%v", ReqRoomList, RspRoomList)
		client.SendMessageToClient([]interface{}{RspRoomList})

	case ReqTableList:
		client.RspRoomTables()

	case ReqJoinRoom:
		if len(data) < 2 {
			logs.Error("user [%d] request join room ,but missing room id", client.User.Id)
			return
		}
		var roomId int
		if id, ok := data[1].(float64); ok {
			roomId = int(id)
		}
		RoomManagers.Lock.RLock()
		defer RoomManagers.Lock.RUnlock()
		if room, ok := RoomManagers.Rooms[roomId]; ok {
			client.Room = room
			res := make([][2]int, 0)
			for _, table := range client.Room.Tables {
				if len(table.Clients) < 3 {
					res = append(res, [2]int{int(table.TableId), len(table.Clients)})
				}
			}
			logs.Debug("RepJoinRoom %v res:%v", RspJoinRoom, res)
			client.SendMessageToClient([]interface{}{RspJoinRoom, res})
		}

	case ReqNewTable:
		table := client.Room.NewTable(client)
		table.JoinTable(client)

	case ReqJoinTable:
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
			table.JoinTable(client)
		}
		client.RspRoomTables()
	case ReqDealPoker:
		if client.Table.State == GameEnd {
			client.Ready = true
		}
	case ReqCallScore:
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

		if (score > 0 && score < client.Table.GameManage.MaxCallScore) || score > 3 {
			logs.Error("player[%d] call score[%d] cheat", client.User.Id, score)
			return
		}
		if score > client.Table.GameManage.MaxCallScore {
			client.Table.GameManage.MaxCallScore = score
			client.Table.GameManage.MaxCallScoreTurn = client
		}
		client.IsCalled = true
		callEnd := score == 3 || client.Table.AllCalled()
		userCall := []interface{}{RspCallScore, client.User.Id, score, callEnd}
		for _, c := range client.Table.Clients {
			logs.Debug("ReqCallScore response:%v", userCall)
			c.SendMessageToClient(userCall)
		}
		if callEnd {
			logs.Debug("call score end")
			client.Table.CallEnd()
		}
	case ReqShotPoker:
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
							message := []interface{}{RspShotPoker, client.User.Id, []int{}}
							for _, c := range client.Table.Clients {
								logs.Debug("ReqShotPoker response:%v", message)
								c.SendMessageToClient(message)
							}
							return
						}
					}
				}
				if len(shotPokers) > 0 {
					compareRes, isMulti := ComparePoker(client.Table.GameManage.LastShotPoker, shotPokers)
					if client.Table.GameManage.LastShotClient != client && compareRes < 1 {
						logs.Warn("player[%d] shot poker %v small than last shot poker %v ", client.User.Id, shotPokers, client.Table.GameManage.LastShotPoker)
						message := []interface{}{RspShotPoker, client.User.Id, []int{}}
						for _, c := range client.Table.Clients {
							logs.Debug("ReqShotPoker response:%v", message)
							c.SendMessageToClient(message)
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
				message := []interface{}{RspShotPoker, client.User.Id, shotPokers}
				logs.Debug("ReqShotPoker response:%v", message)
				for _, c := range client.Table.Clients {
					c.SendMessageToClient(message)
				}
				if len(client.HandPokers) == 0 {
					client.Table.GameOver(client)
				}
			}
		}

	//case ReqGameOver:
	case ReqChat:
		if len(data) > 1 {
			switch data[1].(type) {
			case string:
				client.Table.Chat(client, data[1].(string))
			}
		}
	case ReqRestart:
		client.Table.Lock.Lock()
		defer client.Table.Lock.Unlock()

		if client.Table.State == GameEnd {
			client.Ready = true
			for _, c := range client.Table.Clients {
				if c.Ready == false {
					return
				}
			}
			logs.Debug("restart")
			client.Table.Reset()
		}
	}
}
