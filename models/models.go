package models

import "time"

type Player struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	CurrentBalance uint `json:"currentBalance"`
}
//@TODO использовать как историю изменений баланса
type Balance struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Amount int `json:"amount"`
	Reason string `json:"reason"`
}

type Tournament struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Status string `json:"status"`
	Deposit uint `json:"deposit"`
}

type TournamentParticipant struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	PlayerId uint `json:"playerId"`
	TournamentId uint `json:"tournamentId"`
	BackerIds []uint `gorm:"type:int[]" json:"backerIds"`
}
