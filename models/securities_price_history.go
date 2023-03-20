package models

import (
	"time"
)

type DayPriceModel struct {
	Date                time.Time `json:"date"`
	High                float32   `json:"high"`
	Low                 float32   `json:"low"`
	Open                float32   `json:"open"`
	Close               float32   `json:"close"`
	LTP                 float32   `json:"ltp"`
	PrevClose           float32   `json:"prev_close"`
	TotalTradedQuantity int32     `json:"total_traded_quantity"`
	TotalTradedValue    float64   `json:"total_traded_value"`
	High52W             float32   `json:"high_52w"`
	Low52W              float32   `json:"low_52w"`
	TotalTrades         int32     `json:"total_trades"`
	DeliveryQuantity    int32     `json:"delivery_quantity"`
	DeliveryPercentage  float32   `json:"delivery_percentage"`
}

type SecuritiesPriceHistoryModel struct {
	Symbol  string          `json:"symbol"`
	History []DayPriceModel `json:"history"`
}
