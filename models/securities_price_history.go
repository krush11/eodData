package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type DayPriceModelList []DayPriceModel

func (l DayPriceModelList) Value() (driver.Value, error) {
	// nil slice? produce NULL
	if l == nil {
		return nil, nil
	}
	// empty slice? produce empty array
	if len(l) == 0 {
		return []byte{'{', '}'}, nil
	}

	out := []byte{}
	out = append(out, '{')
	for _, v := range l {
		// This assumes that the date field in the pg composite
		// type accepts the default time.Time format. If that is
		// not the case then you simply provide v.Date in such a
		// format which the composite's field understand, e.g.,
		// v.Date.Format("<layout that the pg composite understands>")
		x := fmt.Sprintf(`"(%v,%v,%v,%v,%v,%v,%v,%v,%f,%v,%v,%v,%v,%v)",`,
			v.Date.Format("2006-01-02"), v.High, v.Low, v.Open, v.Close,
			v.LTP, v.PrevClose, v.TotalTradedQuantity,
			v.TotalTradedValue, v.High52W, v.Low52W, v.TotalTrades,
			v.DeliveryQuantity, v.DeliveryPercentage)
		out = append(out, x...)
	}

	out[len(out)-1] = '}' // replace last "," with "}"
	return out, nil
}

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
	Symbol  string            `json:"symbol"`
	History DayPriceModelList `json:"history"`
}
