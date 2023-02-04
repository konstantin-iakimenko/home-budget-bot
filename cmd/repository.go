package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"math/big"
)

const (
	UserSelect     = "SELECT id FROM users WHERE user_name = $1"
	UserInsert     = "INSERT INTO users(user_name, first_name, last_name, lang) VALUES ($1, $2, $3, $4) RETURNING id"
	BillInsert     = "INSERT INTO bills(user_id, bought_at, description, category, amount, currency, amount_rub, amount_usd) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	BillItemInsert = "INSERT INTO bill_items(bill_id, title, price, cnt, amount, currency, amount_rub, amount_usd) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	CategorySelect = "SELECT category FROM desc_categories WHERE description = $1"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetCategoryByDescription(ctx context.Context, description string) (string, error) {
	var category string
	desc, err := r.pool.Query(ctx, CategorySelect, description)
	defer desc.Close()
	if err != nil {
		return "", err
	}
	if desc.Next() {
		err = desc.Scan(&category)
		if err != nil {
			return "", err
		}
	} else {
		category = "-"
	}
	return category, nil
}

func (r *Repository) SaveBill(ctx context.Context, user *tgbotapi.User, bill *Bill, currency *Currency, usd *Currency) error {
	tx, err := r.pool.BeginTx(
		ctx,
		pgx.TxOptions{
			IsoLevel:       pgx.ReadCommitted,
			AccessMode:     pgx.ReadWrite,
			DeferrableMode: pgx.Deferrable})
	if err != nil {
		return err
	}

	var userId int64
	users, err := tx.Query(ctx, UserSelect, user.UserName)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if users.Next() {
		err = users.Scan(&userId)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	} else {
		err := tx.QueryRow(ctx, UserInsert, user.UserName, user.FirstName, user.LastName, user.LanguageCode).Scan(&userId)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}
	users.Close()

	rubAmount := convertToRub(bill.TotalAmount, currency)
	usdAmount := convertToUsd(bill.TotalAmount, currency, usd)
	var billId int64
	rows, err := tx.Query(ctx, BillInsert, userId, bill.BoughtAt, bill.Description, bill.Category, bill.TotalAmount, currency.NumCode, rubAmount, usdAmount)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	rows.Next()
	err = rows.Scan(&billId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	rows.Close()

	for _, item := range bill.Items {
		rubAmountItem := convertToRub(item.Sum, currency)
		usdAmountItem := convertToUsd(item.Sum, currency, usd)
		_, err = tx.Exec(ctx, BillItemInsert, billId, item.Name, item.Price, item.Count, item.Sum, currency.NumCode, rubAmountItem, usdAmountItem)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	return err
}

func convertToUsd(amount int64, currency *Currency, usd *Currency) int64 {
	if currency.Code == "USD" {
		return amount
	}
	rub := convertToRub(amount, currency)
	usdValue := new(big.Float).Quo(big.NewFloat(float64(rub)), usd.ExRate)
	res, _ := usdValue.Int64()
	return res
}

func convertToRub(amount int64, currency *Currency) int64 {
	if currency.Code == "RUB" {
		return amount
	}
	mul := new(big.Float).Mul(currency.ExRate, big.NewFloat(float64(amount)))
	res, _ := mul.Int64()
	return res
}
