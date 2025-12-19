package main

import (
	"fmt"
	"log"
	"net/http"

	"hilgardvr.com/power-bitcoin-go/internal"
)

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	data, err := internal.GetBitcoinData(app.Environment.ApiBaseUrl, app.Environment.CoinMarketCapKey, false)
	if err != nil {
		log.Println("Error getting bitcoin data", err)
	}
	w.Write([]byte(fmt.Sprintf("%s: %f", data.Data.BtcData.Name, data.Data.BtcData.Quote.Usd.Price)))
}

