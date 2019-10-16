package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	parse "Rusfencing_Telegram_bot/Parse"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type ratingParams struct {
	category    string
	sex, weapon string
}

type kbData struct {
	text         string
	callbackData string
}

var resMap map[int]*parse.Compet
var ratingParMap map[int]*ratingParams
var lastMsg map[int]int

var weapons = []kbData{
	{text: "Сабля", callbackData: "476"},
	{text: "Шпага", callbackData: "475"},
	{text: "Рапира", callbackData: "474"},
}
var s = []kbData{
	{text: "Мужской", callbackData: "450"},
	{text: "Женский", callbackData: "451"},
}
var ages = []kbData{
	{text: "Кадеты", callbackData: "495"},
	{text: "Юниоры", callbackData: "496"},
	{text: "Взрослые", callbackData: "498"},
}

func main() {
	http.HandleFunc("/", MainHandler)
	go func() {
		_ = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("RFgBot")) // TestBotKey RFgBot
	if err != nil {
		log.Fatal("Key err: ", err)
	}
	bot.Debug = true
	log.Printf("Auth on account %s", bot.Self.UserName)
	resMap = make(map[int]*parse.Compet)
	//u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60
	//updates, err := bot.GetUpdatesChan(u)
	ratingParMap = make(map[int]*ratingParams)
	updates := bot.ListenForWebhook("/" + bot.Token)
	for update := range updates {
		go func(update tgbotapi.Update) {
			if update.CallbackQuery != nil {
				var query = update.CallbackQuery.Data
				if query == "495" || query == "496" || query == "498" {
					ratingParMap[update.CallbackQuery.From.ID].category = query
				}
				if query == "450" || query == "451" {
					ratingParMap[update.CallbackQuery.From.ID].sex = query
				}
				if query == "476" || query == "475" || query == "474" {
					ratingParMap[update.CallbackQuery.From.ID].weapon = query
				}
				if ratingParMap[update.CallbackQuery.From.ID].weapon != "" && ratingParMap[update.CallbackQuery.From.ID].category != "" && ratingParMap[update.CallbackQuery.From.ID].sex != "" {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "кнопки нажаты")
					_, _ = bot.Send(msg)
				}
			}
			if update.Message == nil {
				return
			}
			all := make([]string, 0)
			isRating := false
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			if cmd := update.Message.Command(); cmd != "" {
				switch cmd {
				case "results":
					all = []string{getAllCompsResults() + "\nВведите номер турнира, результат которого вам интересен"}
				case "rating":
					all = []string{"В процессе разработки"}
					msg.ReplyToMessageID = update.Message.MessageID
					if update.Message.From.UserName == "AndVl1" {
						isRating = true
					}
				}
			} else {
				//uID := update.Message.From.ID
				if update.Message.Text == "/start" {
					all = []string{"Нажмите /results, далее введите номер интересующего турнира"}
				} else if i, err := strconv.Atoi(update.Message.Text); err == nil && i > 0 && i <= 30 {
					if len(resMap) < 5 {
						_ = getAllCompsResults()
					}
					all = getResultByLink(resMap[i-1].Link)
				}
			}
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			for _, str := range all {
				msg.Text = str
				if isRating {
					keyboard := tgbotapi.InlineKeyboardMarkup{}
					var row []tgbotapi.InlineKeyboardButton
					for _, weapon := range weapons {
						row = append(row, tgbotapi.NewInlineKeyboardButtonData(weapon.text, weapon.callbackData))
					}
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					row = []tgbotapi.InlineKeyboardButton{}
					for _, v := range s {
						row = append(row, tgbotapi.NewInlineKeyboardButtonData(v.text, v.callbackData))
					}
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					row = []tgbotapi.InlineKeyboardButton{}
					for _, age := range ages {
						row = append(row, tgbotapi.NewInlineKeyboardButtonData(age.text, age.callbackData))
					}
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					//tgbotapi.NewEditMessageReplyMarkup(update.Message.Chat.ID, mg.MessageID, keyboard)
					msg.ReplyMarkup = keyboard
					ratingParMap[update.Message.From.ID] = &ratingParams{category: "", sex: "", weapon: ""}
				}
				_, _ = bot.Send(msg)
			}
		}(update)
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
	toSend = fmt.Sprintf("<a href=\"rusfencing.ru%s\">Протокол</a>\n\n", link)
	for _, v := range res[:len(res)/3] {
		toSend += fmt.Sprintf("%s. <a href=\"rusfencing.ru%s\">%s</a>\n", v.Place, v.Link, v.Name)
	}
	all = append(all, toSend)
	toSend = ""
	for _, v := range res[len(res)/3 : len(res)*2/3] {
		toSend += fmt.Sprintf("%s. <a href=\"rusfencing.ru%s\">%s</a>\n", v.Place, v.Link, v.Name)
	}
	all = append(all, toSend)
	toSend = ""
	for _, v := range res[len(res)*2/3:] {
		toSend += fmt.Sprintf("%s. <a href=\"rusfencing.ru%s\">%s</a>\n", v.Place, v.Link, v.Name)
	}
	all = append(all, toSend)
	return all
}

func MainHandler(r http.ResponseWriter, _ *http.Request) {
	_, _ = r.Write([]byte("Hi there"))
}
