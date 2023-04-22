package models

type ReqMove struct {
	LandlordHands []string `json:"landlord_hands"`
	PeasantHands  []string `json:"peasant_hands"`
	LandlordMove  []string `json:"landlord_move"`
	PeasantMove   []string `json:"peasant_move"`
	FirstMove     bool     `json:"first_move"`
}

type RspMove struct {
	LandlordHands []string `json:"landlord_hands"`
	PeasantHands  []string `json:"peasant_hands"`
	LandlordMove  []string `json:"landlord_move"`
	PeasantMove   []string `json:"peasant_move"`
	Msg           string   `json:"msg"`
	Code          int      `json:"code"`
}

type MoveRequest struct {
	Hands    []string `json:"hands"`
	Discards []string `json:"discards"`
	Role     int8     `json:"role"`
}

type MoveResponse struct {
	Move string `json:"move"`
	Node string `json:"node"`
	Code int8   `json:"code"`
}

func NewReqMove(LandlordCards []string, PeasantCards []string, LandlordMove []string, PeasantMove []string, FirstMove bool) *ReqMove {
	return &ReqMove{LandlordCards, PeasantCards, LandlordMove, PeasantMove, FirstMove}
}

func NewRespMove(LandlordCards []string, PeasantCards []string, LandlordMove []string, PeasantMove []string, msg string, code int) *RspMove {
	return &RspMove{LandlordCards, PeasantCards, LandlordMove, PeasantMove, msg, code}
}
