package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

type Application struct {
	Environment Environment
	DB          *sql.DB
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
	conn := SetupDB("power-bitcoin.db", "./internal/migrations")
	app := Application{
		Environment: buildEnv(),
		DB:          conn,
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

func SetupDB(dbName string, migrationDir string) *sql.DB {
	conn, err := sql.Open("sqlite3", fmt.Sprintf("./%s", dbName))
	if err != nil {
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		panic(err)
	}
	doMigrations(conn, migrationDir)
	return conn
}

func doMigrations(conn *sql.DB, migrationDir string) error {
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		panic(err)
	}
	migrations := make(map[int]string)
	for _, f := range files {
		spl := strings.Split(f.Name(), ".")
		if len(spl) != 2 || spl[1] != "sql" {
			fmt.Printf("incompatible migration file name: %q", f.Name())
			continue
		} else {
			fmt.Printf("found migration: %q\n", f.Name())
			prefix := strings.Split(spl[0], "_")[0][1:]
			int, err := strconv.Atoi(prefix)
			if err != nil {
				fmt.Printf("incompatible migration file name: %q", f.Name())
				continue
			}
			_, exists := migrations[int]
			if exists {
				panic(fmt.Errorf("Existing migration version found at %q", int))
			}
			migrations[int] = f.Name()
		}
	}
	keys := make([]int, 0, len(migrations))
	for k := range migrations {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, v := range keys {
		migration, err := os.ReadFile(migrationDir + "/" + migrations[v])
		if err != nil {
			log.Fatal(err)
		}
		_, err = conn.Exec(string(migration))
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	fmt.Printf("ran migrations: %q\n", migrations)
	return nil
}
