package service

import (
	"GoServer/models"
	"bytes"
	"encoding/json"
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
	toRobot    chan []interface{} //发送给robot的消息
	toServer   chan []interface{} //robot发送给服务器
}

// 重置状态
func (c *ClientController) reset() {
	c.User.Role = 1
	c.HandPokers = make([]int, 0, 21)
	c.Ready = false
	c.IsCalled = false
}

// 发送房间内已有的牌桌信息
func (c *ClientController) sendRoomTables() {
	res := make([][2]int, 0)
	for _, table := range c.Room.Tables {
		if len(table.TableClients) < 3 {
			res = append(res, [2]int{int(table.TableId), len(table.TableClients)})
		}
	}
	c.sendMsg([]interface{}{RespTableList, res})
}

func (c *ClientController) sendMsg(msg []interface{}) {
	if c.IsRobot {
		c.toRobot <- msg
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
func (c *ClientController) close() {
	if c.Table != nil {
		for _, client := range c.Table.TableClients {
			if c.Table.Creator == c && c != client {
				c.Table.Creator = client
			}
			if c == client.Next {
				client.Next = nil
			}
		}
		if len(c.Table.TableClients) != 1 {
			for _, client := range c.Table.TableClients {
				if client != client.Table.Creator {
					client.Table.Creator.Next = client
				}
			}
		}
		if len(c.Table.TableClients) == 1 {
			c.Table.Creator = nil
			delete(c.Room.Tables, c.Table.TableId)
			return
		}
		delete(c.Table.TableClients, c.User.Id)
		if c.Table.State == GamePlaying {
			c.Table.syncUser()
			//c.Table.reset()
		}
		if c.IsRobot {
			close(c.toRobot)
			close(c.toServer)
		}
	}
}

func (c *ClientController) readPump() {

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
			logs.Error("message unmarsha1 err, user_id[%d] err:%v", c.User.Id, err)
		} else {
			wsRequest(data, c)
		}
	}
}

// 心跳
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

	go client.readPump()
	go client.Ping()
}
