package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"time"
)

//https://api.telegram.org/botgetMe/387654221:AAHpOV5QcGTzG3CQIUso9SgjOonVKxyLVSs

var (
	users = map[string]string{}
)

func main() {

	bot, err := tgbotapi.NewBotAPI("387654221:AAHpOV5QcGTzG3CQlUso9SgjOonVKxyLVSs")

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//log.Printf("----[%s]: %s", update.Message.From.UserName, update.Message.Text)
		fmt.Println("\n**************************************")
		fmt.Printf("------ from: %+v\n", update.Message.From)
		fmt.Printf("------ contact: %+v\n", update.Message.Contact)
		fmt.Printf("------ chatid: %+v\n", update.Message.Chat.ID)
		fmt.Printf("------ text: %+v\n", update.Message.Text)
		fmt.Println("\n**************************************")

		go func() {
			n := 0
			for {
				n++

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("hello %d", n))
				//msg.ReplyToMessageID = update.Message.MessageID

				ret, err := bot.Send(msg)
				fmt.Println("send ", n, err, ret)
				time.Sleep(time.Second * 10)
			}
		}()
	}
}
