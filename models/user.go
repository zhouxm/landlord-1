package models

import (
	"errors"
	"time"
)

var (
	UserList map[int]*Account
)

func init() {
	UserList = make(map[int]*Account)
	u := Account{1, "77095729@qq.com", "Derek", "password", 1, 0, time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05")}
	UserList[1] = &u
}

type Account struct {
	Id         int    `json:"id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Coin       int    `json:"coin"`
	Role       int    `json:"role"`
	CreateDate string `json:"create_date"`
	UpdateDate string `json:"update_date"`
}

func AddUser(u Account) int {
	u.Id = int(len(UserList) + 1)
	UserList[u.Id] = &u
	return u.Id
}

func GetUser(uid int) (u *Account, err error) {
	if u, ok := UserList[uid]; ok {
		return u, nil
	}
	return nil, errors.New("account not exists")
}

func GetUserByName(username string) (u *Account, err error) {
	for _, u := range UserList {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("account not exists")
}

func GetAllUsers() map[int]*Account {
	return UserList
}

func Login(username, password string) bool {
	for _, u := range UserList {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}
