package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	"strconv"
	"regexp"
	"time"
)

type Player struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	CurrentBalance uint `json:"currentBalance"`
}

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
	Deposit int `json:"deposit"`
}

type TournamentParticipant struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	PlayerId uint `json:"playerId"`
	TournamentId uint `json:"tournamentId"`
	BackerIds []uint `json:"backerIds"`
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
		rest.Get("/joinTournament", i.JoinTournament),
		rest.Get("/resultTournament", i.ResultTournament),
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
	//i.DB.AutoMigrate(&TournamentParticipant{})
	//i.DB.AutoMigrate(&Tournament{})
}

func (i *Impl) InitDB() {
	var err error
	//TODO Переназвать таблицы по-нормальному
	/*gorm.DefaultTableNameHandler = func (db *gorm.DB, defaultTableName string) string  {
		return "prefix_" + defaultTableName;
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

	id := ParseLeadingInt(r.Form.Get("playerId"));
	points := ParseLeadingInt(r.Form.Get("points"));

	player := Player{}

	if i.DB.First(&player, id).Error != nil {
		playerIn := Player{ID: id , CurrentBalance: points};
		i.DB.Create(playerIn);
		i.DB.Save(&playerIn);
		w.WriteJson(playerIn);

		return
	}

	player.CurrentBalance += points;
	i.DB.Save(&player)
	w.WriteJson(player)
}

func (i *Impl) PlayerTake(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	id := ParseLeadingInt(r.Form.Get("playerId"));
	points := ParseLeadingInt(r.Form.Get("points"));

	player := Player{}

	if i.DB.First(&player, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	if player.CurrentBalance - points <= 0 {
		rest.Error(w, "Unlucky boys", 403)
		return
	}

	player.CurrentBalance -= points;
	i.DB.Save(&player)
	w.WriteJson(&player)
}

func (i *Impl) PlayerBalance(w rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	id := ParseLeadingInt(r.Form.Get("playerId"));
	player := Player{}

	if i.DB.First(&player, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	w.WriteJson(&player)}

func (i *Impl) ResetDB(w rest.ResponseWriter, r *rest.Request) {
	i.DB.DropTableIfExists(&Player{}, &Balance{}, &Tournament{}, &TournamentParticipant{})
	i.InitSchemas()
	w.WriteJson("Tables dropped")
}

func (i *Impl) JoinTournament(w rest.ResponseWriter, r *rest.Request) {
	rest.NotFound(w, r)
}

func (i *Impl) ResultTournament(w rest.ResponseWriter, r *rest.Request) {
	rest.NotFound(w, r)
}