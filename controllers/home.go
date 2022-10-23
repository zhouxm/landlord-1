package controllers

import (
	"encoding/json"

	"github.com/beego/beego/v2/core/logs"

	"github.com/beego/beego/v2/server/web"
)

type HomeController struct {
	web.Controller
}

func (c HomeController) Get() {
	c.Layout = "poker.html"
	c.TplName = "poker.html"

	userId := c.Ctx.GetCookie("userid")

	username := c.Ctx.GetCookie("username")
	logs.Debug("user request Index - userId [%v]", userId)
	logs.Debug("user request Index - username [%v]", username)
	user := make(map[string]interface{})
	user["uid"] = userId
	user["username"] = username
	res, err := json.Marshal(user)
	if err != nil {
		logs.Error("user request Index - json marsha1 user %v ,err:%v", user, err)
		return
	}
	c.Data["user"] = string(res)
	c.Data["port"] = "8081"
	err = c.Render()
	c.Ctx.WriteString("PutUsers")
	if err != nil {
		return
	}
}
