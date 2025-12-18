package client

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"fmt"
	"encoding/json"
)

func GetBitcoinData() (CoinMarketCapResponse, error) {
	var response CoinMarketCapResponse
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest", nil)
	if err != nil {
		log.Print(err)
		return response, err
	}

	q := url.Values{}
	q.Add("slug", "bitcoin")

	req.Header.Set("Accepts", "application/json")
	// req.Header.Add("X-CMC_PRO_API_KEY", "b54bcf4d-1bca-4e8e-9a24-22ff2c3d462c")
	req.Header.Add("X-CMC_PRO_API_KEY", "65191b78-6c5c-45e1-83e6-8e3ce0627736")
	req.URL.RawQuery = q.Encode()

	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error sending request to server")
	// 	os.Exit(1)
	// }
	// fmt.Println(resp.Status)
	// respBody, _ := io.ReadAll(resp.Body)
	respBody, err := os.ReadFile("temp.json")
	if err != nil {
		fmt.Println("Error reading file", err)
		return response, err
	}
	fmt.Println(string(respBody))
	json.Unmarshal(respBody, &response)
	return response, err
}