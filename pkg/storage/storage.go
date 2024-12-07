// Пакет для работы с БД приложения GoNews.
package storage

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

type DB struct {
	pool *pgxpool.Pool
}
type Post struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

func New() (*DB, error) {
	godotenv.Load(".env")
	user := os.Getenv("USER")
	pass := os.Getenv("PASSWORD")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	dbname := os.Getenv("DBNAME")
	if user == "" || pass == "" || host == "" || port == "" || dbname == "" {
		return nil, errors.New("Не указано подключение к БД")
	}

	connstr := "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + dbname
	pool, err := pgxpool.Connect(context.Background(), connstr)
	if err != nil {
		return nil, err
	}
	db := DB{
		pool: pool,
	}
	return &db, nil
}

func (db *DB) StoreNews(news []Post) error {
	for _, post := range news {
		_, err := db.pool.Exec(context.Background(), `
		INSERT INTO news(title, content, pub_time, link)
		VALUES ($1, $2, $3, $4)`,
			post.Title,
			post.Content,
			post.PubTime,
			post.Link,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) News(n int) ([]Post, error) {
	if n == 0 {
		n = 10
	}
	rows, err := db.pool.Query(context.Background(), `
	SELECT id, title, content, pub_time, link FROM news
	ORDER BY pub_time DESC
	LIMIT $1
	`,
		n,
	)
	if err != nil {
		return nil, err
	}
	var news []Post
	for rows.Next() {
		var p Post
		err = rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.PubTime,
			&p.Link,
		)
		if err != nil {
			return nil, err
		}
		news = append(news, p)
	}
	return news, rows.Err()
}
