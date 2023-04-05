package service

import (
	"GoServer/models"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

type TableId int

const (
	GameWaitting = iota
	GameCallScore
	GamePlaying
	GameEnd
)

type Table struct {
	Lock         sync.RWMutex
	TableId      TableId
	State        int
	Creator      *ClientController
	TableClients map[int]*ClientController
	GameManage   *GameManage
}

func (table *Table) allCalled() bool {
	for _, client := range table.TableClients {
		if !client.IsCalled {
			return false
		}
	}
	return true
}

// 一局结束
func (table *Table) gameOver(client *ClientController) {
	coin := table.Creator.Room.EntranceFee * table.GameManage.MaxCallScore * table.GameManage.Multiple
	table.State = GameEnd
	for _, c := range table.TableClients {
		res := []interface{}{RespGameOver, client.User.Id}
		if client == c {
			res = append(res, coin*2-100)
		} else {
			res = append(res, coin)
		}
		for _, cc := range table.TableClients {
			if cc != c {
				userPokers := make([]int, 0, len(cc.HandPokers)+1)
				userPokers = append(append(userPokers, cc.User.Id), cc.HandPokers...)
				res = append(res, userPokers)
			}
		}
		logs.Debug("gameOver sendMsg:%v", res)
		c.sendMsg(res)
	}
	logs.Debug("table[%d] game over", table.TableId)
}

// 叫分阶段结束
func (table *Table) callEnd() {
	table.State = GamePlaying
	table.GameManage.FirstCallScore = table.GameManage.FirstCallScore.Next
	if table.GameManage.MaxCallScoreTurn == nil || table.GameManage.MaxCallScore == 0 {
		table.GameManage.MaxCallScoreTurn = table.Creator
		table.GameManage.MaxCallScore = 1
		//return
	}
	landLord := table.GameManage.MaxCallScoreTurn
	landLord.User.Role = RoleLandlord
	table.GameManage.Turn = landLord
	for _, poker := range table.GameManage.Pokers {
		landLord.HandPokers = append(landLord.HandPokers, poker)
	}
	res := []interface{}{RespShowPoker, landLord.User.Id, table.GameManage.Pokers}
	for _, c := range table.TableClients {
		logs.Debug("callEnd sendMsg:%v", res)
		c.sendMsg(res)
	}
}

// 客户端加入牌桌
func (table *Table) joinTable(c *ClientController) {
	table.Lock.Lock()
	defer table.Lock.Unlock()
	if len(table.TableClients) > 2 {
		logs.Error("Player[%d] JOIN Table[%d] FULL", c.User.Id, table.TableId)
		return
	}
	logs.Debug("Player:[%v] id [%v] request join table", c.User.Username, c.User.Id)
	if _, ok := table.TableClients[c.User.Id]; ok {
		logs.Error("Player:[%v] id [%v] already in this table", c.User.Username, c.User.Id)
		return
	}

	c.Table = table
	c.Ready = true
	for _, client := range table.TableClients {
		if client.Next == nil {
			client.Next = c
			break
		}
	}
	table.TableClients[c.User.Id] = c
	table.syncUser()
	if len(table.TableClients) == 3 {
		c.Next = table.Creator
		table.State = GameCallScore
		table.dealPoker()
	} else if c.Room.AllowRobot {
		go table.addRobot(c.Room)
		logs.Debug("robot join ok")
	}
}

// 加入机器人
func (table *Table) addRobot(room *Room) {
	robot := fmt.Sprintf("ROBOT-%d", table.getRobotID())
	logs.Debug("robot [%v] join table", robot)
	if len(table.TableClients) < 3 {
		client := &ClientController{
			Room:       room,
			HandPokers: make([]int, 0, 21),
			User: &models.Account{
				Id:       table.getRobotID(),
				Username: robot,
				Coin:     10000,
			},
			IsRobot:  true,
			toRobot:  make(chan []interface{}, 3),
			toServer: make(chan []interface{}, 3),
		}
		go client.runRobot()
		table.joinTable(client)
	}
}

// 生成随机robotID
func (table *Table) getRobotID() (robot int) {
	time.Sleep(time.Microsecond * 10)
	rand.Seed(time.Now().UnixNano())
	robot = rand.Intn(10000)
	table.Lock.RLock()
	defer table.Lock.RUnlock()
	if _, ok := table.TableClients[robot]; ok {
		return table.getRobotID()
	}
	return
}

// 发牌
func (table *Table) dealPoker() {
	logs.Debug("deal poker")
	table.GameManage.Pokers = make([]int, 0)
	for i := 0; i < 54; i++ {
		table.GameManage.Pokers = append(table.GameManage.Pokers, i)
	}
	table.ShufflePokers()
	for i := 0; i < 17; i++ {
		for _, client := range table.TableClients {
			client.HandPokers = append(client.HandPokers, table.GameManage.Pokers[len(table.GameManage.Pokers)-1])
			table.GameManage.Pokers = table.GameManage.Pokers[:len(table.GameManage.Pokers)-1]
		}
	}
	response := make([]interface{}, 0, 3)
	response = append(append(append(response, RespDealPoker), table.GameManage.FirstCallScore.User.Id), nil)
	for _, client := range table.TableClients {
		sort.Ints(client.HandPokers)
		response[len(response)-1] = client.HandPokers
		logs.Debug("dealPoker sendMsg:%v", response)
		client.sendMsg(response)
	}
}

func (table *Table) chat(client *ClientController, msg string) {
	res := []interface{}{RespChat, client.User.Id, msg}
	for _, c := range table.TableClients {
		logs.Debug("chat sendMsg:%v", res)
		c.sendMsg(res)
	}
}

func (table *Table) reset() {
	table.GameManage = &GameManage{
		FirstCallScore:   table.GameManage.FirstCallScore,
		Turn:             nil,
		MaxCallScore:     0,
		MaxCallScoreTurn: nil,
		LastShotClient:   nil,
		Pokers:           table.GameManage.Pokers[:0],
		LastShotPoker:    table.GameManage.LastShotPoker[:0],
		Multiple:         1,
	}
	table.State = GameCallScore
	if table.Creator != nil {
		logs.Debug("reset table.Creator.sendMsg:%v", []interface{}{RespRestart})
		table.Creator.sendMsg([]interface{}{RespRestart})
	}
	for _, c := range table.TableClients {
		c.reset()
	}
	if len(table.TableClients) == 3 {
		table.dealPoker()
	}
}

// ShufflePokers
// 洗牌
func (table *Table) ShufflePokers() {
	logs.Debug("ShufflePokers")
	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := len(table.GameManage.Pokers)
	for i > 0 {
		randIndex := r.Intn(i)
		table.GameManage.Pokers[i-1], table.GameManage.Pokers[randIndex] = table.GameManage.Pokers[randIndex], table.GameManage.Pokers[i-1]
		i--
	}
}

// 同步用户信息
func (table *Table) syncUser() {
	logs.Debug("sync user")
	response := make([]interface{}, 0, 3)
	response = append(append(response, RespJoinTable), table.TableId)
	tableUsers := make([][2]interface{}, 0, 2)
	current := table.Creator
	for i := 0; i < len(table.TableClients); i++ {
		tableUsers = append(tableUsers, [2]interface{}{current.User.Id, current.User.Username})
		current = current.Next
	}
	response = append(response, tableUsers)
	for _, client := range table.TableClients {
		logs.Debug("syncUser sendMsg:%v", response)
		client.sendMsg(response)
	}
}
