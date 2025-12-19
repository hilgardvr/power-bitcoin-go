package internal

import (
	"log"
	"net/http"
	"net/url"
	"io"
	"encoding/json"
	"os"
)

func GetBitcoinData(baseUrl, apiKey string, live bool) (CoinMarketCapResponse, error) {
	client := &http.Client{}
	var response CoinMarketCapResponse
	req, err := http.NewRequest("GET", baseUrl + "/v1/cryptocurrency/quotes/latest", nil)
	if err != nil {
		log.Print(err)
		return response, err
	}

	q := url.Values{}
	q.Add("slug", "bitcoin")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
	req.URL.RawQuery = q.Encode()

	var respBody []byte
	if live {
		resp, innerErr := client.Do(req)
		if innerErr != nil {
			log.Println("Error sending request to server", innerErr)
			return response, innerErr
		}
		respBody, err = io.ReadAll(resp.Body)
	} else {
		respBody, err = os.ReadFile("temp.json")
	}
	if err != nil {
		log.Println("Error reading response", err)
		return response, err
	}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Println("Error unmarshalling response", err)
	}
	return response, err
}