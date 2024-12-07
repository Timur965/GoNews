// Сервер GoNews.
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"GoNews/pkg/api"
	"GoNews/pkg/rss"
	"GoNews/pkg/storage"
)

// конфигурация приложения
type config struct {
	URLS   []string `json:"rss"`
	Period int      `json:"request_period"`
}

func main() {
	db, err := storage.New()
	if err != nil {
		log.Fatal(err)
	}
	api := api.New(db)

	b, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	var config config
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal(err)
	}

	channelPosts := make(chan []storage.Post)
	channelErrs := make(chan error)
	for _, url := range config.URLS {
		go parseURL(url, db, channelPosts, channelErrs, config.Period)
	}

	go func() {
		for posts := range channelPosts {
			db.StoreNews(posts)
		}
	}()

	go func() {
		for err := range channelErrs {
			log.Println("ошибка:", err)
		}
	}()

	err = http.ListenAndServe(":80", api.Router())
	if err != nil {
		log.Fatal(err)
	}
}
func parseURL(url string, db *storage.DB, posts chan<- []storage.Post, errs chan<- error, period int) {
	for {
		news, err := rss.Parse(url)
		if err != nil {
			errs <- err
			continue
		}
		posts <- news
		time.Sleep(time.Minute * time.Duration(period))
	}
}
