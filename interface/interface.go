package _interface

import (
	"github.com/jinzhu/gorm"
	"log"
	"../models"
)

type Impl struct {
	DB *gorm.DB
}

func (i *Impl) InitSchemas() {
	i.DB.AutoMigrate(&models.Player{})
	i.DB.AutoMigrate(&models.Balance{})
	i.DB.AutoMigrate(&models.TournamentParticipant{})
	i.DB.AutoMigrate(&models.Tournament{})
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