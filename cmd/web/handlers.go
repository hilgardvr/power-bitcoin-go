package main

import (
	"database/sql"
	"errors"
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

type Price struct {
	ID        int64
	Price     float64
	CreatedAt int64
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
	price, err := getCachedPrice(app.DB)
	if err != nil {
		price, err = internal.GetBitcoinData(app.Environment.ApiBaseUrl, app.Environment.CoinMarketCapKey, app.Environment.ApiLive, app.DB)
		_, err := savePrice(price, app.DB)
		if err != nil {
			log.Println("failed to persist price cache: ", err)
		}
	}
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
		CurrentPrice:     price,
		PriceProjections: priceProjections(price, int64(yearsToCalc)),
	}
	err = ts.Execute(w, tmplData)
	if err != nil {
		log.Println("Failed to execute template", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
	log.Println("end :: home")
}

func getCachedPrice(db *sql.DB) (float64, error) {
	now := time.Now().Add(time.Minute * -1).UTC().Unix()
	stmt := `
		select id, price, created_at from price
		where created_at > ?;`
	rows, err := db.Query(stmt, now)
	if err != nil {
		return 0, err
	}
	var price Price
	if rows.Next() {
		rows.Scan(&price.ID, &price.Price, &price.CreatedAt)
		log.Println("found next row: ", price)
	} else {
		log.Println("no price record found")
		err = errors.New("no price record found")
	}
	return price.Price, err
}

func savePrice(price float64, db *sql.DB) (int64, error) {
	stmt := `insert into price (price, created_at) values (?, ?)`
	result, err := db.Exec(stmt, price, time.Now().UTC().Unix())
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}
