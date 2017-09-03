package _interface

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/lib/pq"
	"../models"
)

func (i *Impl) UpdatePlayersBalances(playerIds []uint, difference int) {
	//@TODO проверять на ошибку
	i.DB.Model(&models.Player{}).
		Where("id in (?)", playerIds).
		Update("current_balance", gorm.Expr("current_balance + ?", difference))
}

func (i *Impl) PlayerFund(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	id := ParseLeadingInt(r.Form.Get("playerId"))
	points := ParseLeadingInt(r.Form.Get("points"))

	player := models.Player{}

	if i.DB.First(&player, id).Error != nil {
		playerIn := models.Player{ID: id , CurrentBalance: points}
		i.DB.Create(&playerIn)
		w.WriteJson(&playerIn)

		return
	}

	player.CurrentBalance += points
	i.DB.Save(&player)
	w.WriteJson(&player)
}

func (i *Impl) PlayerTake(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	id := ParseLeadingInt(r.Form.Get("playerId"))
	points := ParseLeadingInt(r.Form.Get("points"))

	player := models.Player{}

	if i.DB.First(&player, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	if player.CurrentBalance - points <= 0 {
		rest.Error(w, "Player balance must be gte 0", http.StatusNotAcceptable)
		return
	}

	player.CurrentBalance -= points
	i.DB.Save(&player)
	w.WriteJson(&player)
}

func (i *Impl) PlayerBalance(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	id := ParseLeadingInt(r.Form.Get("playerId"))
	player := models.Player{}

	if i.DB.First(&player, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	w.WriteJson(&player)
}
