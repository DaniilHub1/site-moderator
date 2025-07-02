package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	resultsDB *sql.DB
	dbOnce    sync.Once
)

func InitResultsDB(path string) error {
	var err error
	dbOnce.Do(func() {
		resultsDB, err = sql.Open("sqlite3", path)
		if err != nil {
			return
		}

		_, err = resultsDB.Exec(`
			CREATE TABLE IF NOT EXISTS results (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				domain TEXT,
				category TEXT,
				word TEXT,
				context TEXT,
				llm_flag BOOLEAN,
				llm_explain TEXT,
				screenshot_path TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`)
	})
	return err
}

func SaveResult(domain, category, word, context string, llmFlag bool, llmExplain, screenshotPath string) error {
	if resultsDB == nil {
		return fmt.Errorf("results DB is not initialized")
	}
	_, err := resultsDB.Exec(`
		INSERT INTO results (domain, category, word, context, llm_flag, llm_explain, screenshot_path) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, domain, category, word, context, llmFlag, llmExplain, screenshotPath)
	if err != nil {
		log.Println("Ошибка сохранения результата:", err)
	}
	return err
}
