package models

import (
	"errors"
	"time"
)

var (
	UserList map[int64]*User
)

func init() {
	UserList = make(map[int64]*User)
	u := User{1, "77095729@qq.com", "Derek", "password", 1, time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05")}
	UserList[1] = &u
}

type User struct {
	Id         int64
	Email      string
	Username   string
	Password   string
	Coin       int
	CreateDate string
	UpdateDate string
}

func AddUser(u User) int64 {
	u.Id = int64(len(UserList) + 1)
	UserList[u.Id] = &u
	return u.Id
}

func GetUser(uid int64) (u *User, err error) {
	if u, ok := UserList[uid]; ok {
		return u, nil
	}
	return nil, errors.New("User not exists")
}

func GetUserByName(username string) (u *User, err error) {
	for _, u := range UserList {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("User not exists")
}

func GetAllUsers() map[int64]*User {
	return UserList
}

func UpdateUser(uid int64, uu *User) (a *User, err error) {
	if u, ok := UserList[uid]; ok {
		if uu.Username != "" {
			u.Username = uu.Username
		}
		if uu.Password != "" {
			u.Password = uu.Password
		}
		if uu.Email != "" {
			u.Email = uu.Email
		}
		if uu.Coin != 0 {
			u.Coin = uu.Coin
		}
		if uu.CreateDate != "" {
			u.CreateDate = uu.CreateDate
		}
		if uu.UpdateDate != "" {
			u.UpdateDate = uu.UpdateDate
		}
		return u, nil
	}
	return nil, errors.New("User Not Exist")
}

func Login(username, password string) bool {
	for _, u := range UserList {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}

func DeleteUser(uid int64) {
	delete(UserList, uid)
}
