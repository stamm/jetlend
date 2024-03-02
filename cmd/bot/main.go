package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"log/slog"

	"github.com/stamm/jetlend/pkg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cron "github.com/robfig/cron/v3"
)

var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("action"),
		tgbotapi.NewKeyboardButton("secondary"),
	),
)

func main() {
	slog.Info("Start daemon")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true
	cronDailyStart(bot)
	cronHourlyStart(bot)

	// logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	logger := log.New(os.Stderr, "[tg]: ", log.LstdFlags|log.Lmsgprefix)
	tgbotapi.SetLogger(logger)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 86400

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		reply := "Для старта нужно ввести куку sessionid из браузера. Эта кука очень приватна. Мы не храним введёную куку, поэтому её нужно указывать каждый раз.\nTo start, you must enter the sessionid cookie from your browser. This cookie is very private. We don't store entered cookie. You need to enter cookie sessionid every time."
		isErr, isStart := false, false
		if update.Message.Text != "/start" {
			switch update.Message.Text {
			case "action":
				data := strings.Split(os.Getenv("JETLEND_CFG"), ";")
				cfg := make(map[int64][]string, len(data))
				for _, v := range data {
					s := strings.Split(v, "=")
					n, err := strconv.Atoi(s[0])
					if err != nil {
						continue
					}
					sids := strings.Split(s[1], ",")
					cfg[int64(n)] = sids
				}
				log.Printf("+%v", cfg)
				ctx := context.Background()
				sids, ok := cfg[update.Message.From.ID]
				if ok {
					var have bool
					reply, have, err = pkg.WhatBuy(ctx, sids, false, true)
					if err != nil {
						reply = "Error: " + err.Error()
						isErr = true
					}
					if !have {
						reply = ""
					}
				}
			case "secondary":
				data := strings.Split(os.Getenv("JETLEND_CFG"), ";")
				cfg := make(map[int64][]string, len(data))
				for _, v := range data {
					s := strings.Split(v, "=")
					n, err := strconv.Atoi(s[0])
					if err != nil {
						continue
					}
					sids := strings.Split(s[1], ",")
					cfg[int64(n)] = sids
				}
				log.Printf("+%v", cfg)
				ctx := context.Background()
				sids, ok := cfg[update.Message.From.ID]
				if ok {
					var have bool
					reply, have, err = pkg.SecondaryMarket(ctx, sids, false, true)
					if err != nil {
						reply = "Error: " + err.Error()
						isErr = true
					}
					if !have {
						reply = ""
					}
				}
			default:
				var err error
				cookies := strings.Split(update.Message.Text, ",")
				reply, err = pkg.Run(context.Background(), cookies, false, true)
				if err != nil {
					reply = "Error: " + err.Error()
					isErr = true
				}
			}
		} else {
			isStart = true
		}

		log.Printf("reply: %s", reply)
		if reply == "" {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = keyboard
		if !isErr && !isStart {
			msg.ParseMode = "MarkdownV2"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("can't send message: %s", err)
		}
	}
}

func cronDailyStart(bot *tgbotapi.BotAPI) error {
	logger := log.New(os.Stderr, "[cron]: ", log.LstdFlags|log.Lmsgprefix)
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	schedule := os.Getenv("JETLEND_SCHEDULE")
	if schedule == "" {
		schedule = "0 21 * * *"
	}
	_, err := c.AddFunc(schedule, func() { sendDaily(bot) })
	c.Start()
	return err
}

func cronHourlyStart(bot *tgbotapi.BotAPI) error {
	logger := log.New(os.Stderr, "[cron]: ", log.LstdFlags|log.Lmsgprefix)
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	schedule := "19 6-17 * * *"
	_, err := c.AddFunc(schedule, func() { sendHourly(bot) })
	c.Start()
	return err
}

func sendDaily(bot *tgbotapi.BotAPI) {
	log.Printf("start sending by cron\n")
	data := strings.Split(os.Getenv("JETLEND_CFG"), ";")
	d, ok := os.LookupEnv("JETLEND_DAYS")
	if !ok {
		d = "31"
	}
	days, err := strconv.Atoi(d)
	if err != nil {
		log.Printf("error converting int to string: %s", err)
		days = 31
	}
	cfg := make(map[int64][]string, len(data))
	for _, v := range data {
		s := strings.Split(v, "=")
		n, err := strconv.Atoi(s[0])
		if err != nil {
			continue
		}
		sids := strings.Split(s[1], ",")
		cfg[int64(n)] = sids
	}
	log.Printf("+%v", cfg)
	ctx := context.Background()
	for chatID, sids := range cfg {

		isErr := false
		reply, err := pkg.Run(ctx, sids, false, true)
		if err != nil {
			reply = "Error: " + err.Error()
			isErr = true
		}

		expectMsg, err := pkg.ExpectAmount(ctx, sids, false, days)
		if err != nil {
			reply += "Error: " + err.Error()
			isErr = true
		} else {
			reply += "\n" + expectMsg
		}

		msg := tgbotapi.NewMessage(chatID, reply)
		if !isErr {
			msg.ParseMode = "MarkdownV2"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("error send for chat_id %d: %s", chatID, err)
		}
	}
}

func sendHourly(bot *tgbotapi.BotAPI) {
	log.Printf("start sending by cron\n")
	data := strings.Split(os.Getenv("JETLEND_CFG"), ";")
	cfg := make(map[int64][]string, len(data))
	for _, v := range data {
		s := strings.Split(v, "=")
		n, err := strconv.Atoi(s[0])
		if err != nil {
			continue
		}
		sids := strings.Split(s[1], ",")
		cfg[int64(n)] = sids
	}
	log.Printf("+%v", cfg)
	ctx := context.Background()
	for chatID, sids := range cfg {
		isErr := false
		reply, have, err := pkg.WhatBuy(ctx, sids, false, true)
		if err != nil {
			reply = "Error: " + err.Error()
			isErr = true
		}
		if !have {
			continue
		}

		msg := tgbotapi.NewMessage(chatID, reply)
		if !isErr {
			msg.ParseMode = "MarkdownV2"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("error send for chat_id %d: %s", chatID, err)
		}
	}
}
