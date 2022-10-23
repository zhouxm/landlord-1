package service

import (
	"time"

	"github.com/beego/beego/v2/core/logs"
)

func (c *ClientController) runRobot() {
	for {
		select {
		case msg, ok := <-c.toServer:
			if !ok {
				return
			}
			wsRequest(msg, c)
		case msg, ok := <-c.toRobot:
			if !ok {
				return
			}
			logs.Debug("robot [%v] receive  message %v ", c.User.Username, msg)
			if len(msg) < 1 {
				logs.Error("send to robot [%v],message err ,%v", c.User.Username, msg)
				return
			}
			if act, ok := msg[0].(int); ok {
				//protocolCode := act
				switch act {
				case RespDealPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.FirstCallScore == c {
						c.autoCallScore()
					}
					c.Table.Lock.RUnlock()

				case RespCallScore:
					if len(msg) < 4 {
						logs.Error("ResCallScore msg err:%v", msg)
						return
					}
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c && !c.IsCalled {
						var callEnd bool
						logs.Debug("ResCallScore %t", msg[3])
						if res, ok := msg[3].(bool); ok {
							callEnd = res
						}
						if !callEnd {
							c.autoCallScore()
						}
					}
					c.Table.Lock.RUnlock()

				case RespShotPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()

				case RespShowPoker:
					time.Sleep(time.Second)
					//logs.Debug("robot [%v] role [%v] receive message ResShowPoker turn :%v", c.User.Username, c.User.Role, c.Table.GameManage.Turn.User.Username)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c || (c.Table.GameManage.Turn == nil && c.User.Role == RoleLandlord) {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()
				case RespGameOver:
					c.Ready = true
				}
			}
		}
	}
}

//自动出牌
func (c *ClientController) autoShotPoker() {
	//因为机器人休眠一秒后才出牌，有可能因用户退出而关闭chan
	defer func() {
		err := recover()
		if err != nil {
			logs.Warn("autoShotPoker err : %v", err)
		}
	}()
	logs.Debug("robot [%v] auto-shot poker", c.User.Username)
	shotPokers := make([]int, 0)
	if len(c.Table.GameManage.LastShotPoker) == 0 || c.Table.GameManage.LastShotClient == c {
		shotPokers = append(shotPokers, c.HandPokers[0])
	} else {
		shotPokers = CardsAbove(c.HandPokers, c.Table.GameManage.LastShotPoker)
	}
	float64Pokers := make([]interface{}, 0)
	for _, poker := range shotPokers {
		float64Pokers = append(float64Pokers, float64(poker))
	}
	req := []interface{}{float64(ReqShotPoker)}
	req = append(req, float64Pokers)
	logs.Debug("robot [%v] autoShotPoker %v", c.User.Username, float64Pokers)
	c.toServer <- req
}

//自动叫分
func (c *ClientController) autoCallScore() {
	defer func() {
		err := recover()
		if err != nil {
			logs.Warn("autoCallScore err : %v", err)
		}
	}()
	logs.Debug("robot [%v] autoCallScore", c.User.Username)
	c.toServer <- []interface{}{float64(ReqCallScore), float64(3)}
}
