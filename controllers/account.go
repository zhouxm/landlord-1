package controllers

import (
	"GoServer/models"
	"encoding/json"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

// AccountController about Users
type AccountController struct {
	beego.Controller
}

// GetAll
// @Title GetAll
// @Description get all Users
// @Success 200 {object} models.Account
func (u AccountController) GetAll() {
	users := models.GetAllUsers()
	u.Data["json"] = users
	u.ServeJSON()
}

// Get
// @Title Get
// @Description get user by uid
// @Param	uid		path 	int64	true		"The key for staticblock"
// @Success 200 {object} models.
// @Failure 403 :uid is empty
func (u AccountController) Get() {
	uid, _ := u.GetInt64(":uid")
	if uid != 0 {
		user, err := models.GetUser(uid)
		if err != nil {
			u.Data["json"] = err.Error()
		} else {
			u.Data["json"] = user
		}
	}
	u.ServeJSON()
}

// Register
// @Title Register
// @Description register users
// @Param	body 	models.Account	true		"body for user content"
// @Success 200 {object} models.Account
// @Failure 403 body is empty
func (u AccountController) Register() {
	var account models.Account
	json.Unmarshal(u.Ctx.Input.RequestBody, &account)
	uid := models.AddUser(account)
	u.Ctx.SetCookie("userid", strconv.Itoa(int(uid)), 86400, "/")
	u.Ctx.SetCookie("username", account.Username, 86400, "/")
	u.Data["json"] = map[string]int64{"uid": uid}
	u.ServeJSON()
}

//
//// @Title Update
//// @Description update the user
//// @Param	uid		path 	int64	true		"The uid you want to update"
//// @Param	body		body 	models.	true		"body for user content"
//// @Success 200 {object} models.
//// @Failure 403 :uid is not int
//// @router /:uid [put]
//func (u *AccountController) Put() {
//	uid, _ := u.GetInt64(":uid")
//	if uid != 0 {
//		var user models.Account
//		json.Unmarshal(u.Ctx.Input.RequestBody, &user)
//		uu, err := models.UpdateUser(uid, &user)
//		if err != nil {
//			u.Data["json"] = err.Error()
//		} else {
//			u.Data["json"] = uu
//		}
//	}
//	u.ServeJSON()
//}

// Login
// @Title Login
// @Description Logs user into the system
// @Param	body body 	models.	true		"body for user content"
// @Success 200 {string} login success
// @Failure 403 403 body is empty
func (u AccountController) Login() {
	var user models.Account
	json.Unmarshal(u.Ctx.Input.RequestBody, &user)
	if models.Login(user.Username, user.Password) {
		u.Data["json"] = "login success"
		user, _ := models.GetUserByName(user.Username)
		u.Ctx.SetCookie("userid", strconv.Itoa(int(user.Id)), 86400, "/")
		u.Ctx.SetCookie("username", user.Username, 86400, "/")
	} else {
		u.Data["json"] = "user not exist"
	}
	u.ServeJSON()
}

// Logout
// @Title logout
// @Description Logs out current logged in user session
// @Success 200 {string} logout success
func (u AccountController) Logout() {
	u.Data["json"] = "logout success"
	u.Ctx.SetCookie("userid", u.Ctx.GetCookie("userid"), -1, "/")
	u.Ctx.SetCookie("username", u.Ctx.GetCookie("username"), -1, "/")
	u.ServeJSON()
}
