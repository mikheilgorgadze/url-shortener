package database

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type GeneratedUrl struct {
    ID        int
    ShortCode string
    LongUrl   string
    AddedTime time.Time
}

var db *sql.DB

func InitDB() error {
    var err error
    dbPath := os.Getenv("DB_PATH")

    log.Printf("DB Path: %v", dbPath)

    if dbPath == "" {
        dbPath = "./data/database.db"
    }
    
    err = os.MkdirAll(filepath.Dir(dbPath), 0755)
    if err != nil {
        return err
    }

    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return err
    }

    err = db.Ping()
    if err != nil {
        return err
    }

    return CreateGeneratedUrlsTable() 
}

func CloseDB() error {
    if db != nil {
        db.Close();
    }
    return nil
}

func CreateGeneratedUrlsTable() error {
    query := `CREATE TABLE IF NOT EXISTS generated_urls(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        short_code TEXT NOT NULL UNIQUE,
        long_url TEXT NOT NULL,
        added_time DATETIME NOT NULL 
    )`
    _, err := db.Exec(query)

    return err
}

func InsertURL(generatedUrl GeneratedUrl) (error) {
    query := `INSERT INTO generated_urls(short_code, long_url, added_time)
    values(?, ?, ?)
    `
    _, err := db.Exec(query, generatedUrl.ShortCode, generatedUrl.LongUrl, generatedUrl.AddedTime)

    return err
}

func GetURLByShortCode(shortCode string) (string, error) {
    var longUrl string
    query := `SELECT long_url FROM generated_urls u WHERE u.short_code = ?`

    err := db.QueryRow(query, shortCode).Scan(&longUrl)
    if err == sql.ErrNoRows {
        return "", errors.New("URL Not Found")
    }
    return longUrl, err
}

func ShortCodeExists(shortCode string) (bool, error) {
    var exists bool
    
    query := `SELECT EXISTS(SELECT 1 FROM generated_urls u WHERE u.short_code = ?)`

    err := db.QueryRow(query, shortCode).Scan(&exists)
    return exists, err
}
