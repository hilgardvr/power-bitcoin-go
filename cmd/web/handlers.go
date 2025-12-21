package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"hilgardvr.com/power-bitcoin-go/internal"
)

var exponent = 5.80
var genesisBlock = time.Date(2009, time.January, 3, 0, 0, 0, 0, time.UTC)

func projectionPriceAt(now time.Time) float64 {
	days := now.Sub(genesisBlock).Hours() / 24
	return math.Pow(10, -17) * math.Pow(days, exponent)
}

type TemplateData struct {
	CurrentPrice     float64
	PriceProjections []PriceProjection
}

type PriceProjection struct {
	Year             int64
	ModelPrice       int64
	AnnualisedChange int64
	TotalChange      int64
	NextYearChange   int64
}

func dummyPriceProjections() []PriceProjection {
	return []PriceProjection{
		PriceProjection{
			Year:             0,
			ModelPrice:       110000,
			AnnualisedChange: 30.0,
			TotalChange:      30.0,
			NextYearChange:   29.0,
		},
		PriceProjection{
			Year:             1,
			ModelPrice:       130000,
			AnnualisedChange: 30.0,
			TotalChange:      30.0,
			NextYearChange:   29.0,
		},
		PriceProjection{
			Year:             2,
			ModelPrice:       150000,
			AnnualisedChange: 25.0,
			TotalChange:      50.0,
			NextYearChange:   24.0,
		},
	}
}

func priceProjections(currentPrice float64, count int64) []PriceProjection {
	now := time.Now()
	var projections []PriceProjection

	for i := int64(0); i <= count; i++ {
		thisProjection := projectionPriceAt(now.AddDate(int(i), 0, 0))
		totalNominalChange := thisProjection - currentPrice
		totalPercentChange := totalNominalChange / currentPrice * 100
		nextProjection := projectionPriceAt(now.AddDate(int(i)+1, 0, 0))
		nextYearPercentChange := (nextProjection - thisProjection) / thisProjection * 100
		projection := PriceProjection{
			Year:             i,
			ModelPrice:       int64(thisProjection),
			AnnualisedChange: 10,
			TotalChange:      int64(totalPercentChange),
			NextYearChange:   int64(nextYearPercentChange),
		}
		projections = append(projections, projection)
	}
	sort.Slice(projections, func(i, j int) bool {
		return projections[i].Year < projections[j].Year
	})
	return projections
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
		CurrentPrice: data.Data.BtcData.Quote.Usd.Price,
		// PriceProjections: dummyPriceProjections(),
		PriceProjections: priceProjections(data.Data.BtcData.Quote.Usd.Price, 20),
	}
	err = ts.Execute(w, tmplData)
	if err != nil {
		log.Println("Failed to execute template")
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
	log.Println("end :: home")
}
