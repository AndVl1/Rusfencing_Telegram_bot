package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"strconv"

	parse "Rusfencing_Telegram_bot/Parse"
	firebase "firebase.google.com/go"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type ratingParams struct {
	category    string
	sex, weapon string
}

var (
	weapons      map[string]string
	s            map[string]string
	ages         map[string]string
	resMap       map[int]*parse.Compet
	ratingParMap map[int]*ratingParams
	lastMsg      map[int64]int
)

var client *firestore.Client

func main() {
	ratingParMap = make(map[int]*ratingParams)
	weapons = make(map[string]string)
	s = make(map[string]string)
	ages = make(map[string]string)
	lastMsg = make(map[int64]int)
	weapons["Сабля"] = "476"
	weapons["Шпага"] = "475"
	weapons["Рапира"] = "474"
	s["Мужчины"] = "450"
	s["Женщины"] = "451"
	ages["Кадеты"] = "495"
	ages["Юниоры"] = "496"
	ages["Взрослые"] = "498"
	http.HandleFunc("/", MainHandler)
	go func() {
		_ = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}()
	ctx := context.Background()
	client = initFirestore(ctx)
	defer client.Close()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("RFgBot")) // TestBotKey RFgBot
	if err != nil {
		log.Fatal("Key err: ", err)
	}
	bot.Debug = true
	log.Printf("Auth on account %s", bot.Self.UserName)
	resMap = make(map[int]*parse.Compet)
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
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")
					res := getRating(ratingParams{
						category: ratingParMap[update.CallbackQuery.From.ID].category,
						sex:      ratingParMap[update.CallbackQuery.From.ID].sex,
						weapon:   ratingParMap[update.CallbackQuery.From.ID].weapon,
					})
					_, _ = bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
						ChatID:    update.CallbackQuery.Message.Chat.ID,
						MessageID: lastMsg[update.CallbackQuery.Message.Chat.ID],
					})
					msg.DisableWebPagePreview = true
					msg.ParseMode = "HTML"
					msg.Text = res
					_, _ = bot.Send(msg)
				}
			}
			if update.Message == nil {
				return
			}
			all := make([]string, 0)
			isRating := false
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			go addToDB(ctx, update, client)
			if cmd := update.Message.Command(); cmd != "" {
				switch cmd {
				case "results":
					all = []string{getAllCompsResults() + "\nВведите номер турнира, результат которого вам интересен"}
				case "rating":
					all = []string{"Выбирете интересующие вас параметры"}
					msg.ReplyToMessageID = update.Message.MessageID
					isRating = true
					lastMsg[update.Message.Chat.ID] = update.Message.MessageID
				}
			} else {
				if update.Message.Text == "/start" {
					all = []string{"Нажмите /results, далее введите номер интересующего турнира, /rating - ситуацию с система отбора"}
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
					for k, v := range weapons {
						row = append(row, tgbotapi.NewInlineKeyboardButtonData(k, v))
					}
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					row = []tgbotapi.InlineKeyboardButton{}
					for k, v := range s {
						row = append(row, tgbotapi.NewInlineKeyboardButtonData(k, v))
					}
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					row = []tgbotapi.InlineKeyboardButton{}
					for k, v := range ages {
						row = append(row, tgbotapi.NewInlineKeyboardButtonData(k, v))
					}
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
					//tgbotapi.NewEditMessageReplyMarkup(update.Message.Chat.ID, mg.MessageID, keyboard)
					msg.ReplyMarkup = keyboard
					ratingParMap[update.Message.From.ID] = &ratingParams{category: "", sex: "", weapon: ""}
				}
				m, _ := bot.Send(msg)
				lastMsg[update.Message.Chat.ID] = m.MessageID
			}
		}(update)
	}
}

func initFirestore(ctx context.Context) *firestore.Client {
	//sa := option.WithCredentialsFile(os.Getenv("HOME") + "/firebaseKey.json")
	sa := option.WithCredentialsJSON([]byte(os.Getenv("firebaseKey")))
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Println(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Println(err)
	}
	return client
}

func addToDB(ctx context.Context, update tgbotapi.Update, client *firestore.Client) {
	name := fmt.Sprintf("%s [%d]", update.Message.From.UserName, update.Message.From.ID)
	_, _, err := client.Collection(name).Add(ctx, map[string]interface{}{
		"chatID":     update.Message.Chat.ID,
		"uID":        update.Message.From.ID,
		"name":       fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
		"messageTxt": update.Message.Text,
		"messageID":  update.Message.MessageID,
	})
	if err != nil {
		log.Println("add to firestore: ", err)
	}
}

//func addText(update tgbotapi.Update) string {
//	if update.
//}

func getRating(params ratingParams) string {
	rfg := "rusfencing.ru"
	res := parse.ParseRatings(fmt.Sprintf("/rating.php?AGE=%s&WEAPON=%s&SEX=%s&SEASON=2028839", params.category, params.weapon, params.sex))
	toSend := fmt.Sprintf("<a href=\"%s/rating.php?AGE=%s&WEAPON=%s&SEX=%s&SEASON=2028839\">Ссылка</a>\n", rfg, params.category, params.weapon, params.sex)
	for _, v := range res {
		toSend += fmt.Sprintf("\n%s.<a href=\"rusfencing.ru%s\"> %s		[%s]</a>", v.Place, v.Link, v.Name, v.Points)
	}
	return toSend
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
