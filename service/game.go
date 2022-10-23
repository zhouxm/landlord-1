package service

type GameManage struct {
	Turn             *ClientController
	FirstCallScore   *ClientController //每局轮转
	MaxCallScore     int               //最大叫分
	MaxCallScoreTurn *ClientController
	LastShotClient   *ClientController
	Pokers           []int
	LastShotPoker    []int
	Multiple         int //加倍
}
