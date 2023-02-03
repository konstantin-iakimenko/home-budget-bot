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
	ErrorHandlingLink = "Не удалось обработать ссылку"
	ErrorSavingBill   = "Не удалось сохранить чек"
	ErrorParseBill    = "Не удалось разобрать счет"
	ErrorGetCurrency  = "Не удалось получить курсы валют"
	Done              = "Готово"
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

	for update := range updates {
		if update.Message != nil {
			log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if strings.HasPrefix(update.Message.Text, SufPursGovRs) {
				bill, err := a.handleLink(update.Message.Text)
				if err != nil {
					log.Error().Err(err).Msg("error handling link")
					a.storeMessage(update.Message.Text)
					a.sendMessage(bot, update.Message.Chat.ID, ErrorHandlingLink)
					continue
				}
				c, err := GetCurrencies(bill.BoughtAt)
				if err != nil {
					log.Error().Err(err).Msg("error getting currencies")
					a.storeMessage(update.Message.Text)
					a.sendMessage(bot, update.Message.Chat.ID, ErrorGetCurrency)
					continue
				}
				err = a.Repository.SaveBill(ctx, update.Message.From, bill, &c.Rsd, &c.Usd)
				if err != nil {
					log.Error().Err(err).Msg("error saving bill")
					a.storeMessage(update.Message.Text)
					a.sendMessage(bot, update.Message.Chat.ID, ErrorSavingBill)
					continue
				}
				log.Info().Msg("bill saved")
				a.sendMessage(bot, update.Message.Chat.ID, Done)
			} else {
				splitted := strings.Split(update.Message.Text, " ")
				c, err := GetCurrencies(time.Unix(int64(update.Message.Date), 0))
				if err != nil {
					log.Error().Err(err).Msg("error getting currencies")
					a.storeMessage(update.Message.Text)
					a.sendMessage(bot, update.Message.Chat.ID, ErrorGetCurrency)
					continue
				}
				totalAmount, currency, err := c.parseAmount(splitted[0])
				if err != nil {
					log.Error().Err(err).Msg("error parsing string bill")
					a.sendMessage(bot, update.Message.Chat.ID, ErrorParseBill)
					continue
				}
				bill := &Bill{
					TotalAmount: totalAmount * 100,
					BoughtAt:    time.Unix(int64(update.Message.Date), 0),
					Description: splitted[1],
					Category:    parseCategory(splitted[1]),
				}

				err = a.Repository.SaveBill(ctx, update.Message.From, bill, currency, &c.Usd)
				if err != nil {
					log.Error().Err(err).Msg("error saving bill")
					a.storeMessage(update.Message.Text)
					a.sendMessage(bot, update.Message.Chat.ID, ErrorSavingBill)
					continue
				}
				log.Info().Msg("string bill saved")
				a.sendMessage(bot, update.Message.Chat.ID, Done)
			}
		}
	}
}

func parseCategory(description string) string {
	switch strings.ToLower(description) {
	case "автобус", "поезд":
		return "Транспорт"
	case "парикмахерская":
		return "Красота"
	case "gym":
		return "Спорт"
	case "подшиводежды", "футболка", "футболки", "джинсы", "одежда", "кроссовки", "ботинки", "туфли":
		return "Одежда"
	case "подарок":
		return "подарки"
	case "кафе", "mac", "kfc", "ресторан":
		return "Рестораны"
	case "sbb", "связь", "телефон", "интернет":
		return "Связь"
	case "вода", "лимонад", "сок", "кола":
		return "Продукты"
	case "панда", "элефант":
		return "Дом и ремонт"
	default:
		return "-"
	}
}

func (a *app) sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
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
