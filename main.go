package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"os"
	parse "rfgTgBot/Parse"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TestBotKey"))
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Auth on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
	}

	ps := parse.ParseCompetitions()
	for _, v := range ps {
		fmt.Println(v.Title, v.Link)
	}
}
