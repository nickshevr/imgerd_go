package _interface

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"../models"
	"math"
)

func (i *Impl) AnnounceTournament(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	tournamentId := ParseLeadingInt(r.Form.Get("tournamentId"))
	deposit := ParseLeadingInt(r.Form.Get("deposit"))

	tournament := models.Tournament{ID: tournamentId, Deposit: deposit, Status: "opened"}

	if i.DB.Create(&tournament).Error != nil {
		rest.Error(w, "Already opened", http.StatusNotAcceptable)
		return
	}

	w.WriteJson(&tournament)
}

func (i *Impl) JoinTournament(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	playerId := ParseLeadingInt(r.Form.Get("playerId"))
	tournamentId := ParseLeadingInt(r.Form.Get("tournamentId"))
	backerIdsQuery := r.Form["backerIds"]
	backerIds := make([]uint, len(backerIdsQuery))

	for i, _ := range backerIdsQuery {
		backerIds[i] = ParseLeadingInt(backerIdsQuery[i])
	}

	mainPlayer := models.Player{}
	tournament := models.Tournament{}
	backers := []models.Player{}
	isAlreadyPlayer := models.TournamentParticipant{}

	if i.DB.First(&tournament, tournamentId).Error != nil {
		rest.Error(w, "Tournament not found", http.StatusNotFound)
		return
	}

	if i.DB.First(&mainPlayer, playerId).Error != nil {
		rest.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	if tournament.Status != "opened" {
		rest.Error(w, "Tournament isnt opened", http.StatusNotAcceptable)
		return
	}

	i.DB.Where("player_id = ?", playerId).Where("tournament_id = ?", tournamentId).First(&isAlreadyPlayer)

	if isAlreadyPlayer.ID == playerId {
		rest.Error(w, "You are already take participant in this tournament", http.StatusNotAcceptable)
		return
	}

	//@TODO переписать участок ниже без лишней копипасты
	if len(backerIds) > 0 {
		i.DB.Where("id in (?)", backerIds).Find(&backers)

		if len(backerIds) > len(backers) {
			rest.Error(w, "Backer not found", http.StatusNotFound)
			return
		}

		playersCount := len(backerIds) + 1
		neededAmount := uint(math.Ceil(float64(tournament.Deposit) / float64(playersCount)))

		if mainPlayer.CurrentBalance - neededAmount < 0 {
			rest.Error(w, "Not enought player money", http.StatusNotAcceptable)
			return
		}

		for _, backer := range backers {
			if backer.CurrentBalance - neededAmount < 0 {
				rest.Error(w, "Not enought backer money", http.StatusNotAcceptable)
				return
			}
		}

		mainPlayer.CurrentBalance -= neededAmount
		i.DB.Save(&mainPlayer)

		i.UpdatePlayersBalances(backerIds, -int(neededAmount));

		tournamentParticipant := models.TournamentParticipant{
			PlayerId: mainPlayer.ID,
			TournamentId: tournament.ID,
			BackerIds: backerIds}

		i.DB.Save(&tournamentParticipant)
		w.WriteJson(&tournamentParticipant)

		return
	}

	if mainPlayer.CurrentBalance - tournament.Deposit < 0 {
		rest.Error(w, "Not enougth player money", http.StatusNotAcceptable)
		return
	}

	mainPlayer.CurrentBalance -= tournament.Deposit

	tournamentParticipant := models.TournamentParticipant{
		PlayerId: mainPlayer.ID,
		TournamentId: tournament.ID,
		BackerIds: []uint{}}

	i.DB.Save(&tournamentParticipant)
	i.DB.Save(&mainPlayer)
	w.WriteJson(&tournamentParticipant)
}

func (i *Impl) ResultTournament(w rest.ResponseWriter, r *rest.Request) {
	input := ResultJson{}
	tournament := models.Tournament{}

	if err := r.DecodeJsonPayload(&input); err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if i.DB.First(&tournament, input.TournamentId).Error != nil {
		rest.Error(w, "Tournament not found", http.StatusNotFound)
		return
	}

	tournament.Status = "closed"
	i.DB.Save(&tournament)

	for _, winner := range input.Winners {
		tournamentParticipant := models.TournamentParticipant{}

		if i.DB.Where("tournament_id = ?", input.TournamentId).
			Where("player_id = ?", winner.PlayerId).
			First(&tournamentParticipant).Error != nil {
			rest.Error(w, "TournamentParticipant not found", http.StatusNotFound)
			return
		}

		prize := int(math.Floor(float64(winner.Prize) / float64(len(tournamentParticipant.BackerIds) + 1)))
		i.UpdatePlayersBalances(append(tournamentParticipant.BackerIds, winner.PlayerId), prize)
	}

	w.WriteJson(http.StatusOK)
}
