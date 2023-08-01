package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	_ "github.com/lib/pq"
	"os"
)

var (
	secretCache, _ = secretcache.New(func(cache *secretcache.Cache) {
		cache.CacheItemTTL = 5000000000 // 5s in nanoseconds
	})
)

type credentials struct {
	Password string `json:"password"`
	DbName   string `json:"dbname"`
}

func HandleRequest() (string, error) {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")

	secret, err := secretCache.GetSecretString("postgres-secret")
	if err != nil {
		return "can't get secret", err
	}
	var c credentials
	err = json.Unmarshal([]byte(secret), &c)
	if err != nil {
		return "can't parse secret", err
	}
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, c.Password, c.DbName,
	)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return "can't ping db", err
	}

	fmt.Println("Successfully connected!")
	return "Successfully connected!", nil
}

func main() {
	lambda.Start(HandleRequest)
}
