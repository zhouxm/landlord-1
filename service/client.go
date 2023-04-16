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

// SendRoomTables 发送房间内已有的牌桌信息
func (c *ClientController) SendRoomTables() {
	res := make([][2]int, 0)
	for _, table := range c.Room.Tables {
		if len(table.TableClients) < 3 {
			res = append(res, [2]int{int(table.TableId), len(table.TableClients)})
		}
	}
	c.SendMsg([]interface{}{RepTableList, res})
}

func (c *ClientController) SendMsg(msg []interface{}) {
	if c.IsRobot {
		c.ToRobot <- msg
		return
	}
	msgByte, err := json.Marshal(msg)
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

func (c *ClientController) RecvMessage() {

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

	go client.RecvMessage()
	go client.Ping()
}

func (c *ClientController) RunRobot() {
	for {
		select {
		case msg, ok := <-c.ToServer:
			if !ok {
				return
			}
			WSRequest(msg, c)
		case msg, ok := <-c.ToRobot:
			if !ok {
				return
			}
			logs.Debug("Player [%v] Receive Message %v ", c.User.Username, msg)
			if len(msg) < 1 {
				logs.Error("Player [%v] Receive Message Error, Message: %v", c.User.Username, msg)
				return
			}
			if act, ok := msg[0].(int); ok {
				//protocolCode := act
				switch act {
				case RepDealPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.FirstCallScore == c {
						c.autoCallScore()
					}
					c.Table.Lock.RUnlock()

				case RepCallScore:
					if len(msg) < 4 {
						logs.Error("ResCallScore msg err:%v", msg)
						return
					}
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c && !c.IsCalled {
						var callEnd bool
						logs.Debug("ResCallScore %t", msg[3])
						if res, ok := msg[3].(bool); ok {
							callEnd = res
						}
						if !callEnd {
							c.autoCallScore()
						}
					}
					c.Table.Lock.RUnlock()

				case RepShotPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()

				case RepShowPoker:
					time.Sleep(time.Second)
					//logs.Debug("robot [%v] role [%v] receive message ResShowPoker turn :%v", c.User.Username, c.User.Role, c.Table.GameManage.Turn.User.Username)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c || (c.Table.GameManage.Turn == nil && c.User.Role == RoleLandlord) {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()
				case RepGameOver:
					c.Ready = true
				}
			} else {
				logs.Error("client.ToRobot msg[0] is not int, the message:%v", msg)
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
	req := []interface{}{int(ReqShotPoker)}
	req = append(req, pokers)
	logs.Debug("robotID:%v autoShotPoker :%v", c.User.Id, pokers)
	c.ToServer <- req
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
	c.ToServer <- []interface{}{int(ReqCallScore), int(3)}
}
