package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
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

var BaseURL string = "https://www.nseindia.com"

func ReqConfig() *http.Request {
	req, _ := http.NewRequest("GET", BaseURL, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Host", "www.nseindia.com")
	req.Header.Add("Referer", "https://www.nseindia.com/get-quotes/equity")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	for _, cookie := range res.Cookies() {
		req.AddCookie(cookie)
	}

	cookies := req.Cookies()
	for i := 0; i < len(cookies); i++ {
		for j := i + 1; j < len(cookies); j++ {
			if cookies[i].Name == cookies[j].Name {
				cookies = append(cookies[:j], cookies[j+1:]...)
				j--
			}
		}
	}
	req.Header.Del("Cookie")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return req
}

func FetchHistoricalData(symbol string, from string, to string, series string) (SecurityDataType, error) {
	req := ReqConfig()

	query := req.URL.Query()
	query.Add("symbol", symbol)
	query.Add("from", from)
	query.Add("to", to)
	query.Add("series", "[\""+series+"\"]")
	req.URL.RawQuery = query.Encode()
	req.URL.Path = "/api/historical/cm/equity"

	client := &http.Client{Timeout: 40 * time.Second}
	res, err := client.Do(req)
	defer res.Body.Close()
	var securityData SecurityDataType
	if err != nil {
		return securityData, err
	}

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
