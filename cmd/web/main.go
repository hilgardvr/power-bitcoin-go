package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Application struct {
	Environment Environment
}

type Environment struct {
	CoinMarketCapKey string
	ApiBaseUrl       string
	ApiLive          bool
	ServerPort       string
}

func main() {
	err := loadEnv(".env")
	if err != nil {
		fmt.Println("Failed to load env file: ", err)
	}
	app := Application{
		Environment: buildEnv(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.home)
	mux.Handle("GET /css/output.css", http.StripPrefix("/css", http.FileServer(http.Dir("./ui/css/"))))
	log.Println("Running server on ", app.Environment.ServerPort)
	err = http.ListenAndServe(app.Environment.ServerPort, mux)
	log.Fatal(err)
}

func loadEnv(filename string) error {
	log.Println("start :: loadEnv")
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
	log.Println("end :: loadEnv")
	return scanner.Err()
}

func buildEnv() Environment {
	var apiLive bool
	if defaultNoEnvVar(os.Getenv("apilive"), "true") == "false" {
		log.Println("API integration is turned off")
		apiLive = false
	} else {
		log.Println("API integration is turned on")
		apiLive = true
	}
	return Environment{
		CoinMarketCapKey: defaultNoEnvVar(os.Getenv("apikey"), "b54bcf4d-1bca-4e8e-9a24-22ff2c3d462c"),
		ApiBaseUrl:       defaultNoEnvVar(os.Getenv("apiurl"), "https://sandbox-api.coinmarketcap.com"),
		ServerPort:       defaultNoEnvVar(os.Getenv("serverport"), ":4000"),
		ApiLive:          apiLive,
	}
}

func defaultNoEnvVar(envValue string, defaultValue string) string {
	if envValue == "" {
		return defaultValue
	} else {
		return envValue
	}
}
