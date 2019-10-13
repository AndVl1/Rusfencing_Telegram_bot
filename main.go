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
	resMap := make(map[int]*parse.Compet)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		toSend := ""
		//uID := update.Message.From.ID
		if update.Message.Text == "/results" {
			res := parse.ParseCompetitions()
			for i, v := range res {
				resMap[i] = v
			}
			for k, v := range resMap {
				toSend += fmt.Sprintf("%d : %s", k+1, v.Title)
			}
		}
		msg.Text = toSend
		_, _ = bot.Send(msg)
	}

	//ps := parse.ParseCompetitions()
	//for _, v := range ps {
	//	fmt.Println(v.Title, v.Link)
	//}
}
