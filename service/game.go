package service

type GameManage struct {
	Turn           *ClientController
	FirstCallScore *ClientController //每局轮转
	MaxCallScore   int               //最大叫分

	MaxCallScoreTurn *ClientController
	LastShotClient   *ClientController
	Pokers           []int
	LastShotPoker    []int
	Multiple         int //加倍
}

// type Game struct {
// 	Turn             *ClientController
// 	FirstCallScore   *ClientController //每局轮转
// 	MaxCallScore     int               //最大叫分
// 	MaxCallScoreTurn *ClientController
// 	LastShotClient   *ClientController
// 	players          []Player
// }

// // AllCalled TableClients map[int]*ClientController
// func (game *Game) AllCalled() bool {
// 	for _, player := range game.players {
// 		if !player.IsCalled {
// 			return false
// 		}
// 	}
// 	return true
// }
