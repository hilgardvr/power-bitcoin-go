package main

import (
	"html/template"
	"log"
	"net/http"

	"hilgardvr.com/power-bitcoin-go/internal"
)

type TemplateData struct {
	Price float64
}

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	log.Println("start :: home")
	data, err := internal.GetBitcoinData(app.Environment.ApiBaseUrl, app.Environment.CoinMarketCapKey, app.Environment.ApiLive)
	if err != nil {
		log.Println("Error getting bitcoin data", err)
		return
	}
	ts, err := template.ParseFiles("./ui/html/pages/home.html")
	if err != nil {
		log.Println("Error parsing home.html")
		return
	}
	tmplData := TemplateData{
		Price: data.Data.BtcData.Quote.Usd.Price,
	}
	err = ts.Execute(w, tmplData)
	if err != nil {
		log.Println("Failed to execute template")
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}
