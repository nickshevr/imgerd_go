package _interface

import (
	"github.com/ant0ine/go-json-rest/rest"
	"../models"
)


func (i *Impl) ResetDB(w rest.ResponseWriter, r *rest.Request) {
	i.DB.DropTableIfExists(&models.Player{}, &models.Balance{}, &models.Tournament{}, &models.TournamentParticipant{})
	i.InitSchemas()
	w.WriteJson("Tables dropped")
}
