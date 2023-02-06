package main

import (
	"encoding/xml"
	"math/big"
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

type ValCurs struct {
	xml.Name `xml:"ValCurs"`
	Valute   []Valute `xml:"Valute"`
}

type Valute struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

type Currency struct {
	NumCode int64
	Code    string
	ExRate  *big.Float
	Symbol  string
}

type CurCash struct {
	m map[string](map[string]Currency)
}
