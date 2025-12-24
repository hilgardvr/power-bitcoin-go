package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"hilgardvr.com/power-bitcoin-go/internal"
)

var exponent = 5.80
var genesisBlock = time.Date(2009, time.January, 3, 0, 0, 0, 0, time.UTC)
var yearsToCalc = 64

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
	ModelPrice       float64
	AnnualisedChange float64
	TotalChange      float64
	NextYearChange   float64
}

func priceProjections(currentPrice float64, count int64) []PriceProjection {
	now := time.Now()
	var projections []PriceProjection

	for i := int64(0); i <= count; i++ {
		thisProjection := projectionPriceAt(now.AddDate(int(i), 0, 0))
		totalNominalChange := thisProjection - currentPrice
		totalPercentChange := totalNominalChange / currentPrice * 100
		totalChangeFactor := thisProjection / currentPrice
		nextProjection := projectionPriceAt(now.AddDate(int(i)+1, 0, 0))
		nextYearPercentChange := (nextProjection - thisProjection) / thisProjection * 100
		annualisedPercentChange := annualisedPercentageChange(totalPercentChange, int(i))
		projection := PriceProjection{
			Year:             i,
			ModelPrice:       thisProjection,
			AnnualisedChange: annualisedPercentChange,
			TotalChange:      totalChangeFactor,
			NextYearChange:   nextYearPercentChange,
		}
		projections = append(projections, projection)
	}
	sort.Slice(projections, func(i, j int) bool {
		return projections[i].Year < projections[j].Year
	})
	return projections
}

func annualisedPercentageChange(totalPercentChange float64, years int) float64 {
	if years == 0 {
		return totalPercentChange
	}
	percentFactor := 1 + (totalPercentChange / 100)
	percent := (math.Pow(percentFactor, 1/float64(years))) * 100
	return percent - 100
}

var templateFunctions = template.FuncMap{
	"prettyPrintFloat64": prettyPrintFloat64,
}

func prettyPrintFloat64(number float64) string {
	p := message.NewPrinter(language.English)
	formatted := p.Sprintf("%.2f", number)
	return formatted
}

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	log.Println("start :: home")
	data, err := internal.GetBitcoinData(app.Environment.ApiBaseUrl, app.Environment.CoinMarketCapKey, app.Environment.ApiLive)
	if err != nil {
		log.Println("Error getting bitcoin data", err)
		return
	}
	ts, err := template.New("home.html").Funcs(templateFunctions).ParseFiles("./ui/html/pages/home.html")
	if err != nil {
		log.Println("Error parsing home.html", err)
		return
	}
	tmplData := TemplateData{
		CurrentPrice:     data.Data.BtcData.Quote.Usd.Price,
		PriceProjections: priceProjections(data.Data.BtcData.Quote.Usd.Price, int64(yearsToCalc)),
	}
	err = ts.Execute(w, tmplData)
	if err != nil {
		log.Println("Failed to execute template", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
	log.Println("end :: home")
}
