package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Price struct {
	ID        int64
	Price     float64
	CreatedAt int64
}

func GetBitcoinData(baseUrl, apiKey string, live bool, db *sql.DB) (float64, error) {
	price, err := getCachedPrice(db)
	if err != nil {
		println("failed to get cached price, trying api.. ", err.Error())
	} else {
		println("found price: ", price)
		return price, nil
	}
	client := &http.Client{}
	var response CoinMarketCapResponse
	var respBody []byte
	if live {
		log.Println("Hitting bitcoin price api")
		req, innerErr := http.NewRequest("GET", baseUrl+"/v1/cryptocurrency/quotes/latest", nil)
		if innerErr != nil {
			log.Print(innerErr)
			return response.Data.BtcData.Quote.Usd.Price, innerErr
		}

		q := url.Values{}
		q.Add("slug", "bitcoin")

		req.Header.Set("Accepts", "application/json")
		req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
		req.URL.RawQuery = q.Encode()

		resp, innerErr := client.Do(req)
		if innerErr != nil {
			log.Println("Error sending request to server", innerErr)
			return response.Data.BtcData.Quote.Usd.Price, innerErr
		}
		respBody, err = io.ReadAll(resp.Body)
	} else {
		respBody, err = os.ReadFile("temp.json")
	}
	if err != nil {
		log.Println("Error reading response", err)
		return response.Data.BtcData.Quote.Usd.Price, err
	}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Println("Error unmarshalling response", err)
	}
	return response.Data.BtcData.Quote.Usd.Price, err
}

func getCachedPrice(db *sql.DB) (float64, error) {
	stmt := `select id, price, created_at from price;`
	rows, err := db.Query(stmt)
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
