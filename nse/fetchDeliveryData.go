package nse

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"updateEODData/utils"
)

type DeliveryDataType struct {
	NoBlockDeals   bool `json:"noBlockDeals"`
	BulkBlockDeals []struct {
		Name string `json:"name"`
	} `json:"bulkBlockDeals"`
	MarketDeptOrderBook []interface{} `json:"marketDeptOrderBook"`
	SecurityWiseDP 	 struct {
		QuantityTraded	int32 `json:"quantityTraded"`
		DeliveryQuantity int32 `json:"deliveryQuantity"`
		DeliveryToTradedQuantity float32 `json:"deliveryToTradedQuantity"`
		SeriesRemarks any `json:"seriesRemarks"`
		SecWiseDelPosDate string `json:"secWiseDelPosDate"`
	} `json:"securityWiseDP"`
}

func FetchDeliveryData(symbol string) (DeliveryDataType, error) {
	req := utils.ReqConfig()

	query := req.URL.Query()
	query.Add("symbol", symbol)
	query.Add("section", "trade_info")
	req.URL.RawQuery = query.Encode()
	req.URL.Path = "/api/quote-equity"

	client := &http.Client{Timeout: 40 * time.Second}
	res, err := client.Do(req)
	var deliveryData DeliveryDataType
	if err != nil {
		return deliveryData, err
	}
	defer res.Body.Close()
	
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader io.ReadCloser
		gzipReader, _ = gzip.NewReader(res.Body)
		err = json.NewDecoder(gzipReader).Decode(&deliveryData)
		defer gzipReader.Close()
	case "application/json":
		var results map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&results)
		return deliveryData, err
	default:
		bodyBytes, _ := io.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		return deliveryData, err
	}

	return deliveryData, err
}
