package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Score struct {
	WPM      float64
	Accuracy float64
	Timestamp time.Time
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./texts.db")
	if err != nil {
		log.Fatal(err)
	}

	createTextsTableSQL := `
	CREATE TABLE IF NOT EXISTS texts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL UNIQUE
	);
	`
	_, err = db.Exec(createTextsTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	createScoresTableSQL := `
	CREATE TABLE IF NOT EXISTS scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		wpm REAL NOT NULL,
		accuracy REAL NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(createScoresTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Insert some initial texts if the table is empty
	insertInitialTexts()
}

func insertInitialTexts() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM texts").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		textsToInsert := []string{
			"The early bird catches the worm, but the second mouse gets the cheese.",
			"Innovation distinguishes between a leader and a follower. Stay curious.",
			"The only way to do great work is to love what you do. Find your passion.",
			"Success is not final, failure is not fatal: it is the courage to continue that counts.",
			"Believe you can and you're halfway there. Doubt kills more dreams than failure ever will.",
		}
		for _, text := range textsToInsert {
			_, err := db.Exec("INSERT OR IGNORE INTO texts (content) VALUES (?)", text)
			if err != nil {
				log.Printf("Error inserting text: %v\n", err)
			}
		}
		fmt.Println("Initial texts inserted into the database.")
	}
}

func getRandomTextFromDB() (string, error) {
	var text string
	rand.Seed(time.Now().UnixNano())
	row := db.QueryRow("SELECT content FROM texts ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&text)
	if err != nil {
		return "", fmt.Errorf("failed to get random text from DB: %w", err)
	}
	return text, nil
}

func saveScore(wpm, accuracy float64) error {
	_, err := db.Exec("INSERT INTO scores (wpm, accuracy) VALUES (?, ?)", wpm, accuracy)
	if err != nil {
		return fmt.Errorf("failed to save score: %w", err)
	}
	return nil
}

func getTopScoresFromDB() ([]Score, error) {
	rows, err := db.Query("SELECT wpm, accuracy, timestamp FROM scores ORDER BY wpm DESC, accuracy DESC LIMIT 10")
	if err != nil {
		return nil, fmt.Errorf("failed to get top scores: %w", err)
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var s Score
		if err := rows.Scan(&s.WPM, &s.Accuracy, &s.Timestamp); err != nil {
			log.Printf("Error scanning score row: %v", err)
			continue
		}
		scores = append(scores, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating score rows: %w", err)
	}

	return scores, nil
}