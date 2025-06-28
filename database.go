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
	Username string
	WPM      float64
	Accuracy float64
	CalculatedScore float64
	Difficulty string
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
		username TEXT NOT NULL,
		wpm REAL NOT NULL,
		accuracy REAL NOT NULL,
		calculated_score REAL NOT NULL,
		difficulty TEXT NOT NULL DEFAULT 'Unknown',
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(createScoresTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Check if 'difficulty' column exists and migrate if not
	rows, err := db.Query("PRAGMA table_info(scores);")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columnExists := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt_value sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			log.Fatal(err)
		}
		if name == "difficulty" {
			columnExists = true
			break
		}
	}

	if !columnExists {
		log.Println("Migrating scores table: Adding 'difficulty' column.")
		_, err = db.Exec("ALTER TABLE scores ADD COLUMN difficulty TEXT NOT NULL DEFAULT 'Unknown';")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Migration complete.")
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

func GetRandomTextFromDB() (string, error) {
	rand.Seed(time.Now().UnixNano())
	numSentences := rand.Intn(5) + 2 // Randomly choose between 2 and 6 sentences
	
	var combinedText string
	for i := 0; i < numSentences; i++ {
		var text string
		row := db.QueryRow("SELECT content FROM texts ORDER BY RANDOM() LIMIT 1")
		err := row.Scan(&text)
		if err != nil {
			return "", fmt.Errorf("failed to get random text from DB: %w", err)
		}
		combinedText += text + " "
	}
	return combinedText, nil
}

func SaveScore(username string, wpm, accuracy float64, difficulty string) error {
	calculatedScore := wpm * (accuracy / 100.0) // Calculate score
	_, err := db.Exec("INSERT INTO scores (username, wpm, accuracy, calculated_score, difficulty) VALUES (?, ?, ?, ?, ?)", username, wpm, accuracy, calculatedScore, difficulty)
	if err != nil {
		return fmt.Errorf("failed to save score: %w", err)
	}
	return nil
}

func GetTopScoresFromDB() ([]Score, error) {
	rows, err := db.Query("SELECT username, wpm, accuracy, calculated_score, difficulty, timestamp FROM scores ORDER BY calculated_score DESC LIMIT 10")
	if err != nil {
		return nil, fmt.Errorf("failed to get top scores: %w", err)
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var s Score
		if err := rows.Scan(&s.Username, &s.WPM, &s.Accuracy, &s.CalculatedScore, &s.Difficulty, &s.Timestamp); err != nil {
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