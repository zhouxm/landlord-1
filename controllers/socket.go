package controllers

//
//import (
//	"encoding/json"
//	"net/http"
//
//	"github.com/beego/beego/v2/core/logs"
//	"github.com/beego/beego/v2/server/web"
//	"github.com/gorilla/websocket"
//)
//
//// WebSocketController handles WebSocket requests.
//type WebSocketController struct {
//	web.Controller
//}
//
//// Get method handles GET requests for WebSocketController.
//func (c *WebSocketController) Get() {
//	// Safe check.
//	uname := c.GetString("uname")
//	if len(uname) == 0 {
//		c.Redirect("/", 302)
//		return
//	}
//
//	c.TplName = "websocket.html"
//	c.Data["IsWebSocket"] = true
//	c.Data["UserName"] = uname
//}
//
//// Join method handles WebSocket requests for WebSocketController.
//func (c *WebSocketController) Join() {
//	uname := c.GetString("uname")
//	if len(uname) == 0 {
//		c.Redirect("/", 302)
//		return
//	}
//
//	// Upgrade from http request to WebSocket.
//	ws, err := websocket.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil, 1024, 1024)
//	if _, ok := err.(websocket.HandshakeError); ok {
//		http.Error(c.Ctx.ResponseWriter, "Not a websocket handshake", 400)
//		return
//	} else if err != nil {
//		logs.Error("Cannot setup WebSocket connection:", err)
//		return
//	}
//
//	// Join chat room.
//	Join(uname, ws)
//	defer Leave(uname)
//
//	// Message receive loop.
//	for {
//		_, p, err := ws.ReadMessage()
//		if err != nil {
//			return
//		}
//		publish <- newEvent(models.EVENT_MESSAGE, uname, string(p))
//	}
//}
//
//// broadcastWebSocket broadcasts messages to WebSocket users.
//func broadcastWebSocket(event models.Event) {
//	data, err := json.Marshal(event)
//	if err != nil {
//		logs.Error("Fail to marshal event:", err)
//		return
//	}
//
//	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
//		// Immediately send event to WebSocket users.
//		ws := sub.Value.(Subscriber).Conn
//		if ws != nil {
//			if ws.WriteMessage(websocket.TextMessage, data) != nil {
//				// User disconnected.
//				unsubscribe <- sub.Value.(Subscriber).Name
//			}
//		}
//	}
//}
