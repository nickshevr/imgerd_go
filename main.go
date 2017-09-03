package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/lib/pq"
	"./interface"
)

func main() {
	i := _interface.Impl{}
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