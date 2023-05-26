package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
	"updateEODData/models"
	"updateEODData/nse"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var pool *pgxpool.Pool
var dateToday = time.Now()
var erroneousSymbols []string
var wg sync.WaitGroup
var errorCount int16 = 0
var symbolCount int16 = 0

func ConnectPSQL() {
	var err error
	config, err := pgxpool.ParseConfig(os.Getenv("PSQL_URL"))
	if err != nil {
		log.Println("Error parsing config: ", err)
	}
	config.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		RegisterDataTypes(ctx, c)
		return true
	}
	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Println("Error connecting to database: ", err)
	}

	if err != nil {
		log.Println("Error pinging database: ", err)
	} else {
		log.Println("Ping response successful")
	}
}

func RegisterDataTypes(ctx context.Context, conn *pgx.Conn) error {
	dataTypeNames := []string{
		"equity.day_price",
		"equity.day_price[]",
	}

	for _, typeName := range dataTypeNames {
		dataType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return err
		}
		conn.TypeMap().RegisterType(dataType)
	}

	return nil
}

func updateSymbol(symbol string) {
	// defer wg.Done()
	startTime := time.Now()
	data, err := nse.FetchHistoricalData(symbol, dateToday.Format("02-01-2006"), dateToday.Format("02-01-2006"), "EQ")

	if err != nil {
		log.Println("Error fetching data for", symbol, " on ", dateToday.Format("02-01-2006"))
		log.Panicln(err)
		return
	}
	if len(data.Data) == 0 {
		log.Println("No data found for", symbol, "on", dateToday.Format("02-01-2006"), "Skipping...")
		erroneousSymbols = append(erroneousSymbols, symbol)
		errorCount++
		return
	}
	if data.Data[0].MTimestamp != dateToday.Format("02-Jan-2006") {
		log.Println("No data found today for", symbol, " on ", dateToday.Format("02-Jan-2006"))
		erroneousSymbols = append(erroneousSymbols, symbol)
		errorCount++
		return
	}
	deliveryData, _ := nse.FetchDeliveryData(symbol)

	var SecuritiesPriceHistory models.SecuritiesPriceHistoryModel
	SecuritiesPriceHistory.Symbol = symbol

	var DayPrice models.DayPriceModel
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
	DayPrice.DeliveryQuantity = deliveryData.SecurityWiseDP.DeliveryQuantity
	DayPrice.DeliveryPercentage = deliveryData.SecurityWiseDP.DeliveryToTradedQuantity
	SecuritiesPriceHistory.History = append(SecuritiesPriceHistory.History, DayPrice)

	_, err = pool.Exec(context.Background(), `UPDATE equity.securities_price_history
	SET history = $1::equity.day_price[] || history WHERE symbol = $2`,
		SecuritiesPriceHistory.History, symbol)
	if err != nil {
		log.Println(1, err)
	}

	log.Println("Updated EOD data for", symbolCount, symbol, "in", time.Since(startTime).Seconds(), "seconds")
}

func HandleRequest() {
	log.Println("Starting updating EOD data for", dateToday.Format("02-01-2006"))
	godotenv.Load()
	ConnectPSQL()

	rows, _ := pool.Query(context.Background(), "SELECT symbol FROM equity.securities_metadata ORDER BY symbol ASC")

	for rows.Next() {
		symbolCount++
		var symbol string
		rows.Scan(&symbol)
		log.Println("Symbol:", symbolCount, symbol)

		// wg.Add(1)
		// go updateSymbol(symbol)
		updateSymbol(symbol)
	}
	// wg.Wait()

	log.Println("Completed updating EOD data for", dateToday.Format("02-01-2006"))
	log.Println("Total symbols updated:", symbolCount, "Total symbols skipped:", errorCount)
}

func main() {
	HandleRequest()
}
