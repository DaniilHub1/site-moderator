package db

import (
	"encoding/csv"
	"fmt"
	"os"
)

var allDomains []string
var offset int

func LoadCSV(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("ошибка открытия CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("ошибка чтения CSV: %w", err)
	}

	for i, row := range records {
		if i == 0 {
			continue 
		}
		if len(row) > 0 {
			allDomains = append(allDomains, row[0])
		}
	}

	fmt.Printf("Загружено %d доменов из CSV\n", len(allDomains))
	return nil
}

func GetDomainsBatch(limit int) ([]string, error) {
	if offset >= len(allDomains) {
		return nil, nil 
	}

	end := offset + limit
	if end > len(allDomains) {
		end = len(allDomains)
	}

	batch := allDomains[offset:end]
	offset = end
	return batch, nil
}
