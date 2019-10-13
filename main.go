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
		all := make([]string, 0)
		//uID := update.Message.From.ID
		if update.Message.Text == "/start" {
			all = []string{"Нажмите /results, далее введите номер интересующего турнира"}
		} else if update.Message.Text == "/results" {
			all = []string{getAllCompsResults() + "\nВведите номер турнира, результат которого вам интересен"}
		} else if i, err := strconv.Atoi(update.Message.Text); err == nil {
			all = getResultByLink(resMap[i-1].Link)
		}
		msg.DisableWebPagePreview = true
		msg.ParseMode = "HTML"
		for _, str := range all {
			msg.Text = str
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

func getResultByLink(link string) []string {
	res := parse.ParseResults(link)
	all := make([]string, 0)
	toSend := ""
	if res == nil {
		return []string{fmt.Sprintf("Командные соревнования пока не получается смотреть. Вот вам ссылка: <a href=\"rusfencing.ru%s\">Результат</a>\n", link)}
	}

	for _, v := range res[:len(res)/2] {
		toSend += fmt.Sprintf("%s. <a href=\"rusfencing.ru%s\">%s</a>\n", v.Place, v.Link, v.Name)
	}
	all = append(all, toSend)
	toSend = ""
	for _, v := range res[len(res)/2:] {
		toSend += fmt.Sprintf("%s. <a href=\"rusfencing.ru%s\">%s</a>\n", v.Place, v.Link, v.Name)
	}
	all = append(all, toSend)
	return all
}
