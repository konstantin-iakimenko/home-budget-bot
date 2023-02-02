package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	UserSelect     = "SELECT id FROM users WHERE user_name = $1"
	UserInsert     = "INSERT INTO users(user_name, first_name, last_name, lang) VALUES ($1, $2, $3, $4) RETURNING id"
	BillInsert     = "INSERT INTO bills(user_id, bought_at, description, category, amount) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	BillItemInsert = "INSERT INTO bill_items(bill_id, title, price, cnt, amount) VALUES ($1, $2, $3, $4, $5)"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveBill(ctx context.Context, user *tgbotapi.User, bill *Bill) error {
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
		users.Close()
	} else {
		err := tx.QueryRow(ctx, UserInsert, user.UserName, user.FirstName, user.LastName, user.LanguageCode).Scan(&userId)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	var billId int64
	rows, err := tx.Query(ctx, BillInsert, userId, bill.BoughtAt, bill.Description, bill.Category, bill.TotalAmount)
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
		_, err := tx.Exec(ctx, BillItemInsert, billId, item.Name, item.Price, item.Count, item.Sum)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	return err
}
