package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

const SufPursGovRs = "https://suf.purs.gov.rs/"

const (
	ErrorHandlingLink    = "Не удалось обработать ссылку"
	ErrorSavingBill      = "Не удалось сохранить чек"
	ErrorParsingBill     = "Не удалось разобрать счет"
	ErrorGettingCategory = "Не удалось получить категорию"
	ErrorGettingCurrency = "Не удалось получить курс валюты"
	Done                 = "Готово"
)

type app struct {
	Repository *Repository
}

func (a *app) Serve(ctx context.Context) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("HOMEBUDGET_BOT_ID"))
	if err != nil {
		log.Error().Err(err).Msg("error creating bot")
		panic(err)
	}
	log.Info().Msgf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	curCash := InitCurCash()

	for update := range updates {
		if update.Message != nil {
			log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if strings.HasPrefix(update.Message.Text, SufPursGovRs) {
				bill, err := a.handleLink(update.Message.Text)
				if err != nil {
					a.sendErrMessage(err, ErrorHandlingLink, bot, update)
					continue
				}
				rsd, err := curCash.Get(bill.BoughtAt, "RSD")
				if err != nil {
					a.sendErrMessage(err, ErrorGettingCurrency, bot, update)
					continue
				}
				usd, err := curCash.Get(bill.BoughtAt, "USD")
				if err != nil {
					a.sendErrMessage(err, ErrorGettingCurrency, bot, update)
					continue
				}
				err = a.Repository.SaveBill(ctx, update.Message.From, bill, rsd, usd)
				if err != nil {
					a.sendErrMessage(err, ErrorSavingBill, bot, update)
					continue
				}
				log.Info().Msg("bill saved")
				a.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, Done)
			} else {
				splitted := strings.Split(update.Message.Text, " ")
				totalAmount, currency, err := parseAmount(splitted[0], curCash, time.Unix(int64(update.Message.Date), 0))
				if err != nil {
					a.sendErrMessage(err, ErrorParsingBill, bot, update)
					continue
				}

				category, err := a.Repository.GetCategoryByDescription(ctx, splitted[1])
				if err != nil {
					a.sendErrMessage(err, ErrorGettingCategory, bot, update)
					continue
				}

				bill := &Bill{
					TotalAmount: totalAmount * 100,
					BoughtAt:    time.Unix(int64(update.Message.Date), 0),
					Description: splitted[1],
					Category:    category,
				}

				usd, err := curCash.Get(bill.BoughtAt, "USD")
				if err != nil {
					a.sendErrMessage(err, ErrorGettingCurrency, bot, update)
					continue
				}
				err = a.Repository.SaveBill(ctx, update.Message.From, bill, currency, usd)
				if err != nil {
					a.sendErrMessage(err, ErrorSavingBill, bot, update)
					continue
				}
				log.Info().Msg("string bill saved")
				a.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, Done)
			}
		}
	}
}

func (a *app) sendErrMessage(err error, errMsg string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Error().Err(err).Msg(errMsg)
	a.storeMessage(update.Message.Text)
	a.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, errMsg)
}

func (a *app) sendMessage(bot *tgbotapi.BotAPI, chatID int64, messageId int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	msg.ReplyToMessageID = messageId
	if err != nil {
		log.Error().Stack().Err(err).Msg("error sending message")
	}
}

func (a *app) storeMessage(message string) {
	f, err := os.OpenFile("errorLinks", os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = os.Create("errorLinks")
			f, err = os.OpenFile("errorLinks", os.O_APPEND|os.O_WRONLY, os.ModePerm)
			if err != nil {
				log.Error().Stack().Err(err).Msg("error saving message on disc")
				return
			}
		} else {
			log.Error().Stack().Err(err).Msg("error saving message on disc")
			return
		}
	}
	defer func() { _ = f.Close() }()

	if _, err = f.WriteString(message + "\n"); err != nil {
		log.Error().Stack().Err(err).Msg("error saving message on disc")
	}
}
