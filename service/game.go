package service

import (
	"landlord/service/agent"
)

type GameManage struct {
	Turn             *agent.ClientController
	FirstCallScore   *agent.ClientController //每局轮转
	MaxCallScore     int                     //最大叫分
	MaxCallScoreTurn *agent.ClientController
	LastShotClient   *agent.ClientController
	Pokers           []int
	LastShotPoker    []int
	Multiple         int //加倍
}
