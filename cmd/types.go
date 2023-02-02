package main

import (
	"time"
)

type Bill struct {
	TotalAmount int64
	BoughtAt    time.Time
	Description string
	Category    string
	Items       []Item
}

type Item struct {
	Name  string
	Price int64
	Count float64
	Sum   int64
}
