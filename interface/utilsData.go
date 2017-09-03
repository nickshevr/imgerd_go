package _interface

import (
	"strconv"
	"regexp"
)

var leadingInt = regexp.MustCompile(`^[-+]?\d+`)

func ParseLeadingInt(s string) (uint) {
	s = leadingInt.FindString(s)
	if s == "" {
		return 0
	}

	res, _ := strconv.ParseUint(s, 10, 64)

	return uint(res)
}

type Winners struct {
	PlayerId uint `json:"playerId"`
	Prize uint `json:"prize"`
}

type ResultJson struct {
	TournamentId uint `json:"tournamentId"`
	Winners []Winners `json:"winners"`

}
