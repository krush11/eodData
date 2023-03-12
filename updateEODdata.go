package main

import (
	"context"
	"log"
	"os"
	"time"
	"updateEODData/models"
	"updateEODData/nse"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var pool *pgxpool.Pool

func ConnectPSQL() {
	var err error
	config, err := pgxpool.ParseConfig(os.Getenv("PSQL_URL"))
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

func RegisterDataTypes(ctx context.Context) error {
	dataTypeNames := []string{
		"equity.day_price",
		"equity.day_price[]",
	}
	conn, _ := pool.Acquire(ctx)
	defer conn.Release()

	for _, typeName := range dataTypeNames {
		dataType, err := conn.Conn().LoadType(ctx, typeName)
		if err != nil {
			return err
		}
		conn.Conn().TypeMap().RegisterType(dataType)
	}

	return nil
}

func HandleRequest() {
	dateToday := time.Now().AddDate(0, 0, -2)
	log.Println("Starting updating EOD data for", dateToday.Format("02-01-2006"))
	godotenv.Load()
	ConnectPSQL()

	err := RegisterDataTypes(context.Background())
	if err != nil {
		log.Println("Error registering data types: ", err)
	}

	rows, _ := pool.Query(context.Background(), "SELECT symbol FROM equity.securities_metadata ORDER BY symbol ASC")

	i := 0
	j := 0
	var erroneousSymbols []string
	for rows.Next() {
		i++
		var symbol string
		rows.Scan(&symbol)
		log.Println("Symbol:", i, symbol)

		startTime := time.Now()
		data, err := nse.FetchHistoricalData(symbol, dateToday.Format("02-01-2006"), dateToday.Format("02-01-2006"), "EQ")

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
		log.Println(SecuritiesPriceHistory.History[0])

		_, err = pool.Exec(context.Background(), `UPDATE equity.securities_price_history
		SET history = $1::equity.day_price[] || history WHERE symbol = $2`,
			SecuritiesPriceHistory.History, "0")
		if err != nil {
			log.Println(err)
		}

		sphm := &models.SecuritiesPriceHistoryModel{}
		err = pool.QueryRow(context.Background(),
			`SELECT (symbol, history) FROM equity.securities_price_history WHERE symbol=$1`,
			"PAYTM").Scan(sphm)
		if err != nil {
			log.Println(err)
		}

		history := []models.DayPriceModel{
			{time.Now().AddDate(0, 0, -2), 4, 1, 2, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{time.Now().AddDate(0, 0, -1), 10, 5, 6, 7, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		}

		insertStmt := `INSERT INTO equity.securities_price_history VALUES ($1, $2)`
		_, err = pool.Exec(context.Background(), insertStmt, "01", history)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Updated EOD data for", i, symbol, "in", time.Since(startTime).Seconds(), "seconds")

		return
	}

	log.Println("Completed updating EOD data for", dateToday.Format("02-01-2006"))
	log.Println("Total symbols updated:", i, "Total symbols skipped:", j)
}

func main() {
	HandleRequest()
}
