package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"strings"
	"bufio"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest", nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := url.Values{}
	q.Add("slug", "bitcoin")

	req.Header.Set("Accepts", "application/json")
	// req.Header.Add("X-CMC_PRO_API_KEY", "b54bcf4d-1bca-4e8e-9a24-22ff2c3d462c")
	req.Header.Add("X-CMC_PRO_API_KEY", app.environment.CoinMarketCapKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}
	fmt.Println(resp.Status)
	respBody, err := io.ReadAll(resp.Body)
	// respBody, err := os.ReadFile("temp.json")
	if err != nil {
		fmt.Println("Error reading file", err)
	}
	fmt.Println(string(respBody))
	var response CoinMarketCapResponse
	// err = json.NewDecoder(r.Body).Decode(&response)
	json.Unmarshal(respBody, &response)
	if err != nil {
		fmt.Println("Error parsing respoinse", err)
	} else {
		fmt.Println(response)
	}
	w.Write([]byte(fmt.Sprintf("%s: %f", response.Data.BtcData.Name, response.Data.BtcData.Quote.Usd.Price)))
}

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func buildEnv() Environment {
	return Environment {
		CoinMarketCapKey: os.Getenv("apikey"),
	}
}

func main() {
	err := loadEnv(".env")
	if err != nil {
		fmt.Println("Failed to load env file: ", err)
	}
	app := application {
		environment: buildEnv(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.home)
	err = http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}

type application struct {
	environment Environment
}

type Environment struct {
	CoinMarketCapKey string
}


type CoinMarketCapResponse struct {
    Status CoinMarketCapResponseStatus `json:"status"`
    Data CoinMarketCapResponseDataEntry `json:"data"`
}

type CoinMarketCapResponseStatus struct {
    Timestamp time.Time `json:"timestamp"`
    ErrorCode int64 `json:"error_code"`
    ErrorMessage string `json:"error_message"`
    Elapsed int64 `json:"elapsed"`
    CreditCount int64 `json:"credit_count"`
    Notice string `json:"notice"`
}

type CoinMarketCapResponseDataEntry struct {
	BtcData CoinMarketCapResponseData `json:"1"`
}

type CoinMarketCapResponseData struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Symbol string `json:"symbol"`
	Slug string `json:"slug"`
	MaxSupply string `json:"max_supply"`
	LastUpdated time.Time `json:"last_updated:"`
	Quote CoinMarketCapCurrencyQuote `json:"quote"`
}

type CoinMarketCapCurrencyQuote struct {
	Usd CoinMarketCapQuote `json:"USD"`
}

type CoinMarketCapQuote struct {
	Price float64 `json:"price"`
	MarketCap float64 `json:"market_cap"`
	MarketCapDominance float64 `json:"market_cap_dominance"`
	LastUpdated time.Time `json:"last_updated"`
}