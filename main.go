package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"os"
	parse "rfgTgBot/Parse"
	"strconv"
)

var resMap map[int]*parse.Compet

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TestBotKey"))
	if err != nil {
		log.Fatal("Key err: ", err)
	}
	bot.Debug = true
	log.Printf("Auth on account %s", bot.Self.UserName)
	resMap = make(map[int]*parse.Compet)
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
			toSend = getAllCompsResults()
		} else if i, err := strconv.Atoi(update.Message.Text); err != nil {
			toSend = getResultByLink(resMap[i].Link)
		}
		msg.Text = toSend
		_, _ = bot.Send(msg)
		msg.Text = "\n\nВведите номер турнира, результат которого вам интересен"
	}
}

func getAllCompsResults() string {
	res := parse.ParseCompetitions()
	toSend := ""
	for i, v := range res {
		resMap[i] = v
	}
	for i := 0; i < len(resMap); i++ {
		toSend += fmt.Sprintf("%d: %s %s\n\n", i+1, resMap[i].Title, resMap[i].Categs)
	}
	return toSend
}

func getResultByLink(link string) string {
	res := parse.ParseResults(link)
	toSend := ""
	for i, v := range res {
		toSend += fmt.Sprintf("%d: %s (rusfencing.ru%s)\n", i+1, v.Name, v.Link)
	}
	return toSend
}
