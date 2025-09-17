package main

import (
  "fmt"
  "tgbot/config"
  "tgbot/handler/message"

  "github.com/amarnathcjd/gogram/telegram"
)

const (
    botToken = "7078343610:AAGDn3AT6gAKAW_BOTIzla-Ciw4RFYjq2bw"
    phoneNumber = "+22892888812"
)

func main() {
    cfg, err := config.NewConfig("database/config.json")
    if err != nil {
        panic(err)
    }

    userBot, _ := telegram.NewClient(telegram.ClientConfig{
        AppID:    cfg.ApiID,
        AppHash:  cfg.ApiHash,
        Session: "database/user.session",
    })
    if err = userBot.Connect(); err != nil {
        panic(err)
    }
    if _, err := userBot.Login(phoneNumber); err != nil {
        panic(err)
    }
    uBot, err := userBot.GetMe()
    if err != nil {
        panic(err)
    }
    bot, _ := telegram.NewClient(telegram.ClientConfig{
        AppID:    cfg.ApiID,
        AppHash:  cfg.ApiHash,
        Session: "database/bot.session",
    })
    if err := bot.Connect(); err != nil {
        panic(err)
    }
    bot.LoginBot(botToken)
    bots, err := bot.GetMe()
    if err != nil {
        panic(err)
    }
    message.RegisterMessageHandler(bot, userBot, cfg)

    fmt.Printf("Userbot -> Logged in as: @%s\n", uBot.Username)
    fmt.Printf("Bot -> Logged in as: @%s\n", bots.Username)

    go bot.Idle()
    userBot.Idle()
}
