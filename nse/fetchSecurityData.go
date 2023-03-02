package nse

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
	"updateEODData/utils"
)

type DailyDataType struct {
	ID                  string  `json:"_id"`
	Symbol              string  `json:"CH_SYMBOL"`
	Series              string  `json:"CH_SERIES"`
	MarkeyType          string  `json:"CH_MARKET_TYPE"`
	High                float32 `json:"CH_TRADE_HIGH_PRICE"`
	Low                 float32 `json:"CH_TRADE_LOW_PRICE"`
	Open                float32 `json:"CH_OPENING_PRICE"`
	Close               float32 `json:"CH_CLOSING_PRICE"`
	LTP                 float32 `json:"CH_LAST_TRADED_PRICE"`
	PrevClose           float32 `json:"CH_PREVIOUS_CLS_PRICE"`
	TotalTradedQuantity int32   `json:"CH_TOT_TRADED_QTY"`
	TotalTradedValue    float64 `json:"CH_TOT_TRADED_VAL"`
	High52W             float32 `json:"CH_52WEEK_HIGH_PRICE"`
	Low52W              float32 `json:"CH_52WEEK_LOW_PRICE"`
	TotalTrades         int32   `json:"CH_TOTAL_TRADES"`
	ISIN                string  `json:"CH_ISIN"`
	ChTimestamp         string  `json:"CH_TIMESTAMP"`
	Timestamp           string  `json:"TIMESTAMP"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
	V                   int8    `json:"__v"`
	VWAP                float32 `json:"VWAP"`
	MTimestamp          string  `json:"mTIMESTAMP"`
}

type SecurityDataType struct {
	Meta interface{}     `json:"meta"`
	Data []DailyDataType `json:"data"`
}

func FetchHistoricalData(symbol string, from string, to string, series string) (SecurityDataType, error) {
	req := utils.ReqConfig()

	query := req.URL.Query()
	query.Add("symbol", symbol)
	query.Add("from", from)
	query.Add("to", to)
	query.Add("series", "[\""+series+"\"]")
	req.URL.RawQuery = query.Encode()
	req.URL.Path = "/api/historical/cm/equity"

	client := &http.Client{Timeout: 40 * time.Second}
	var securityData SecurityDataType
	res, err := client.Do(req)
	if err != nil {
		return securityData, err
	}
	defer res.Body.Close()

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader io.ReadCloser
		gzipReader, _ = gzip.NewReader(res.Body)
		err = json.NewDecoder(gzipReader).Decode(&securityData)
		defer gzipReader.Close()
	case "application/json":
		var results map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&results)
	default:
		bodyBytes, _ := io.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		return securityData, errors.New(bodyString)
	}

	if err != nil {
		return securityData, err
	}

	return securityData, err
}
