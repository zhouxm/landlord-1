package controllers

import (
	"encoding/json"
	beego "github.com/beego/beego/v2/server/web"
	"landlord/endgame"
	"landlord/models"
)

type EndgameController struct {
	beego.Controller
}

// Move
// @Title Get Move
// @Description Get Move
// @Param	body body 	models.ReqMove	true	"The ReqMove content"
// @Success 200 {object} models.RspMove
// @Failure 403 body is empty
// @router /move [post]
func (c *EndgameController) Move() {
	var reqMove models.ReqMove
	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &reqMove)
	result, _ := endgame.Move(&reqMove)
	c.Data["json"] = result
	c.Ctx.Output.Status = result.Code
	//c.ServeJSON()
}
