package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
)

//https://api.telegram.org/botgetMe/387654221:AAHpOV5QcGTzG3CQIUso9SgjOonVKxyLVSs

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

	n := 0
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		n++
		fmt.Printf("chat: %+v", update.Message.Chat)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("hello %d", n))
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
