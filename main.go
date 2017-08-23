package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	"strconv"
	"regexp"
	"time"
	_ "github.com/lib/pq"
	"math"
)

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

func main() {
	i := Impl{}
	i.InitDB()
	i.InitSchemas()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/fund", i.PlayerFund),
		rest.Get("/take", i.PlayerTake),
		rest.Get("/resetDB", i.ResetDB),
		rest.Get("/balance", i.PlayerBalance),
		rest.Get("/announceTournament", i.AnnounceTournament),
		rest.Get("/joinTournament", i.JoinTournament),
		rest.Post("/resultTournament", i.ResultTournament),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":3042", api.MakeHandler()))
}



type Impl struct {
	DB *gorm.DB
}

var leadingInt = regexp.MustCompile(`^[-+]?\d+`)


//@TODO дописать функции парсинга для ID (> 0)
func ParseLeadingInt(s string) (uint) {
	s = leadingInt.FindString(s)
	if s == "" {
		return 0
	}

	res, _ := strconv.ParseUint(s, 10, 64)

	return uint(res)
}

func (i *Impl) InitSchemas() {
	i.DB.AutoMigrate(&Player{})
	i.DB.AutoMigrate(&Balance{})
	i.DB.AutoMigrate(&TournamentParticipant{})
	i.DB.AutoMigrate(&Tournament{})
}

func (i *Impl) UpdatePlayersBalances(playerIds []uint, difference int) {
	//@TODO проверять на ошибку
	i.DB.Model(&Player{}).
		Where("id in (?)", playerIds).
		Update("current_balance", gorm.Expr("current_balance + ?", difference))
}

func (i *Impl) InitDB() {
	var err error
	//TODO Переназвать таблицы по-нормальному
	/*gorm.DefaultTableNameHandler = func (db *gorm.DB, defaultTableName string) string  {
		return "prefix_" + defaultTableName
	}*/

	i.DB, err = gorm.Open("postgres", "host=localhost user=test dbname=gostyle sslmode=disable password=test")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}
}

func (i *Impl) PlayerFund(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	id := ParseLeadingInt(r.Form.Get("playerId"))
	points := ParseLeadingInt(r.Form.Get("points"))

	player := Player{}

	if i.DB.First(&player, id).Error != nil {
		playerIn := Player{ID: id , CurrentBalance: points}
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

	player := Player{}

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
	player := Player{}

	if i.DB.First(&player, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	w.WriteJson(&player)
}

func (i *Impl) ResetDB(w rest.ResponseWriter, r *rest.Request) {
	i.DB.DropTableIfExists(&Player{}, &Balance{}, &Tournament{}, &TournamentParticipant{})
	i.InitSchemas()
	w.WriteJson("Tables dropped")
}

func (i *Impl) AnnounceTournament(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	tournamentId := ParseLeadingInt(r.Form.Get("tournamentId"))
	deposit := ParseLeadingInt(r.Form.Get("deposit"))

	tournament := Tournament{ID: tournamentId, Deposit: deposit, Status: "opened"}

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
	//@TODO запарсить из query в массив
	backerIds := r.Form.Get("backerIds")

	mainPlayer := Player{}
	tournament := Tournament{}
	backers := []Player{}
	isAlreadyPlayer := TournamentParticipant{}

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

		playersCount := uint(len(backerIds) + 1)
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

		//i.UpdatePlayersBalances(backerIds, neededAmount);

		//@TODO цикл, конечно, отстой, можно 1 запросом
		for _, backer := range backers {
			backer.CurrentBalance -= neededAmount
			i.DB.Save(&backer)
		}

		tournamentParticipant := TournamentParticipant{
			PlayerId: mainPlayer.ID,
			TournamentId: tournament.ID,
			//BackerIds: []uint(backerIds)}
			BackerIds: []uint{}}

		i.DB.Save(&tournamentParticipant)

		return
	}

	if mainPlayer.CurrentBalance - tournament.Deposit < 0 {
		rest.Error(w, "Not enougth player money", http.StatusNotAcceptable)
		return
	}

	mainPlayer.CurrentBalance -= tournament.Deposit

	tournamentParticipant := TournamentParticipant{
		PlayerId: mainPlayer.ID,
		TournamentId: tournament.ID,
		BackerIds: []uint{}}

	i.DB.Save(&tournamentParticipant)
	i.DB.Save(&mainPlayer)
	w.WriteJson(&tournamentParticipant)
}

type Winners struct {
	PlayerId uint `json:"playerId"`
	Prize uint `json:"prize"`
}

type ResultJson struct {
	TournamentId uint `json:"tournamentId"`
	Winners []Winners `json:"winners"`

}


func (i *Impl) ResultTournament(w rest.ResponseWriter, r *rest.Request) {
	input := ResultJson{}
	tournament := Tournament{}

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
		tournamentParticipant := TournamentParticipant{}

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