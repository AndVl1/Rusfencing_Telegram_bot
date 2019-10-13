package main

import (
	parse "Rusfencing_Telegram_bot/Parse"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"os"
	"strconv"
)

var resMap map[int]*parse.Compet

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("RFgBot"))
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
		toSend2 := ""
		//uID := update.Message.From.ID
		if update.Message.Text == "/start" {
			toSend = "Нажмите /results, далее введите номер интересующего турнира"
		} else if update.Message.Text == "/results" {
			toSend = getAllCompsResults() + "\nВведите номер турнира, результат которого вам интересен"
		} else if i, err := strconv.Atoi(update.Message.Text); err == nil {
			toSend, toSend2 = getResultByLink(resMap[i-1].Link)
		}
		msg.DisableWebPagePreview = true
		msg.Text = toSend
		msg.ParseMode = "HTML"
		_, _ = bot.Send(msg)
		if toSend2 != "" {
			msg.Text = toSend2
			_, _ = bot.Send(msg)
		}
	}
}

func getAllCompsResults() string {
	res := parse.ParseCompetitions()
	toSend := ""
	for i, v := range res {
		resMap[i] = v
	}
	for i := 0; i < len(resMap); i++ {
		toSend += fmt.Sprintf("%d. %s %s\n\n", i+1, resMap[i].Title, resMap[i].Categs)
	}
	return toSend
}

func getResultByLink(link string) (string, string) {
	res := parse.ParseResults(link)
	toSend := ""
	toSend2 := ""
	if res == nil {
		return fmt.Sprintf("Командные соревнования пока не получается смотреть. Вот вам ссылка: <a href=\"rusfencing.ru%s\">Результат</a>\n", link), ""
	}

	for i, v := range res[:len(res)/2] {
		toSend += fmt.Sprintf("%d. <a href=\"rusfencing.ru%s\">%s</a>\n", i+1, v.Link, v.Name)
	}
	for i, v := range res[len(res)/2:] {
		toSend2 += fmt.Sprintf("%d. <a href=\"rusfencing.ru%s\">%s</a>\n", len(res)/2+i+1, v.Link, v.Name)
	}
	return toSend, toSend2
}
