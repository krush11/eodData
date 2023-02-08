package main

import (
	"context"
	"log"
	"os"
	"time"

	// "github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var db1, db2 *pgx.Conn

func ConnectPSQL() {
	var err error
	db1, err = pgx.Connect(context.Background(), os.Getenv("PSQL_URL"))
	db2, err = pgx.Connect(context.Background(), os.Getenv("PSQL_URL"))
	if err != nil {
		log.Println("Error connecting to database: ", err)
	}

	err = db1.Ping(context.Background())
	err = db2.Ping(context.Background())
	if err != nil {
		log.Println("Error pinging database: ", err)
	} else {
		log.Println("Ping response successful")
	}
}

func HandleRequest() {
	dateToday := time.Now()
	dateYesterday := time.Now().AddDate(0, 0, -1)
	log.Println("Starting updating EOD data for", dateToday.Format("02-01-2006"))
	godotenv.Load()
	ConnectPSQL()

	rows, _ := db1.Query(context.Background(), "SELECT symbol FROM equity.securities_price_history ORDER BY symbol ASC")

	i := 0
	j := 0
	var erroneousSymbols []string
	for rows.Next() {
		i++
		var symbol string
		rows.Scan(&symbol)
		if symbol < "KAYNES" {
			continue
		} 
		startTime := time.Now()
		data, err := FetchHistoricalData(symbol, dateYesterday.Format("02-01-2006"), dateToday.Format("02-01-2006"), "EQ")
		if err != nil {
			log.Println("Error fetching data for", symbol, " on ", dateToday.Format("02-01-2006"))
			log.Panicln(err)
			continue
		}
		if len(data.Data) == 0 {
			log.Println("No data found for", symbol, "on", dateToday.Format("02-01-2006"), "Skipping...")
			erroneousSymbols = append(erroneousSymbols, symbol)
			j++
			continue
		}
		if data.Data[0].MTimestamp != dateToday.Format("02-Jan-2006") {
			log.Println("No data found today for", symbol, " on ", dateToday.Format("02-Jan-2006"))
			erroneousSymbols = append(erroneousSymbols, symbol)
			j++
			continue
		}

		var SecuritiesPriceHistory SecuritiesPriceHistoryModel
		SecuritiesPriceHistory.Symbol = symbol

		var DayPrice DayPriceModel
		DayPrice.Date, _ = time.Parse("02-Jan-2006", data.Data[0].MTimestamp)
		DayPrice.High = data.Data[0].High
		DayPrice.Low = data.Data[0].Low
		DayPrice.Open = data.Data[0].Open
		DayPrice.Close = data.Data[0].Close
		DayPrice.LTP = data.Data[0].LTP
		DayPrice.PrevClose = data.Data[0].PrevClose
		DayPrice.TotalTradedQuantity = data.Data[0].TotalTradedQuantity
		DayPrice.TotalTradedValue = data.Data[0].TotalTradedValue
		DayPrice.High52W = data.Data[0].High52W
		DayPrice.Low52W = data.Data[0].Low52W
		DayPrice.TotalTrades = data.Data[0].TotalTrades
		SecuritiesPriceHistory.History = append(SecuritiesPriceHistory.History, DayPrice)

		db2.Exec(context.Background(), `UPDATE equity.securities_price_history
		SET history = $1::equity.day_price[] || history WHERE symbol = $2`,
			SecuritiesPriceHistory.History, SecuritiesPriceHistory.Symbol)

		log.Println("Updated EOD data for", i, symbol, "in", time.Since(startTime).Seconds(), " seconds")
	}

	log.Println("Completed updating EOD data for", dateToday.Format("02-01-2006"))
	log.Println("Total symbols updated:", i, "Total symbols skipped:", j)
}

func main() {
	// lambda.Start(HandleRequest)
	HandleRequest()
}
