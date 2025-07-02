package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"site-moderator/internal/analyzer"
	"site-moderator/internal/db"
	"site-moderator/internal/fetcher"
	"site-moderator/internal/llm"
	"site-moderator/internal/screenshot"
)

func main() {
	fmt.Println(" Site Moderator запущен")

	err := analyzer.LoadKeywords("keywords.json")
	if err != nil {
		log.Fatal("Не удалось загрузить keywords.json:", err)
	}

	err = db.LoadCSV("domains.csv")
	if err != nil {
		log.Fatal("Ошибка загрузки CSV:", err)
	}

	err = db.InitResultsDB("results.sqlite")
	if err != nil {
		log.Fatal("Ошибка инициализации базы результатов:", err)
	}

	for {
		
		domains, _ := db.GetDomainsBatch(5)
		if len(domains) == 0 {
			fmt.Println("Все домены обработаны.")
			break
		}

		for _, domain := range domains {
			url := addProtocol(domain)
			fmt.Println("🌐 Обработка:", url)

			text, err := fetcher.CrawlSite(url, 3)
			if err != nil {
				fmt.Println("⚠️ Ошибка обхода:", err)
				continue
			}

			fmt.Printf("📄 Собрано текста: %d символов\n", len(text))


			resultAnalyzer, foundAnalyzer := analyzer.CheckText(text)
			if foundAnalyzer {
				fmt.Printf("🚨 Нарушение (анализатор): %s\n", resultAnalyzer.Category)
				fmt.Printf("🔎 Найдено слово: \"%s\"\n", resultAnalyzer.Word)
				fmt.Printf("🧠 Контекст: ...%s...\n", strings.TrimSpace(resultAnalyzer.Context))

				if resultAnalyzer.IsSuspected {
					fmt.Println("⚠️ Контекст похож на описание фильма (не точно)")
				}
			} else {
				fmt.Println("✅ Нарушений (анализатор) не обнаружено")
			}

			// Проверка через нейросеть (LLM)
			resultLLM, foundLLM := llm.CheckWithLLM(text)

			if foundLLM {
				if resultLLM.HasViolation {
					fmt.Printf("🤖 Нарушение (LLM): %s\n", resultLLM.Category)
					fmt.Printf("🔎 Найдено слово: \"%s\"\n", resultLLM.Word)
					fmt.Printf("🧠 Контекст: ...%s...\n", strings.TrimSpace(resultLLM.Context))

					if resultLLM.IsSuspected {
						fmt.Println("⚠️ Контекст похож на описание фильма (не точно)")
					}
				} else {
					fmt.Println("✅ Нарушений (LLM) не обнаружено")
				}
			} else {
				fmt.Println("⚠️ LLM не дал результата")
			}

			if (foundAnalyzer && resultAnalyzer.Category != "") || (foundLLM && resultLLM.HasViolation) {
				savePath := filepath.Join("screenshots", domain+".png")
				err := screenshot.TakeScreenshot(url, savePath)
				if err != nil {
					fmt.Println("⚠️ Ошибка скрина:", err)
				} else {
					fmt.Println("📸 Скриншот сохранён:", savePath)
				}

				category := ""
				word := ""
				context := ""
				hasViolation := false
				explanation := ""
				imagePath := savePath

				if foundLLM && resultLLM.HasViolation {
					category = resultLLM.Category
					word = resultLLM.Word
					context = resultLLM.Context
					hasViolation = true
					if resultLLM.IsSuspected {
						explanation = "Подозрение на описание фильма"
					}
				} else if foundAnalyzer && resultAnalyzer.Category != "" {
					category = resultAnalyzer.Category
					word = resultAnalyzer.Word
					context = resultAnalyzer.Context
					hasViolation = true
					if resultAnalyzer.IsSuspected {
						explanation = "Подозрение на описание фильма"
					}
				}

				err = db.SaveResult(domain, category, word, context, hasViolation, explanation, imagePath)
				if err != nil {
					fmt.Println("⚠️ Ошибка сохранения результата:", err)
				}
			} else {
				err = db.SaveResult(domain, "", "", "", false, "", "")
				if err != nil {
					fmt.Println("⚠️ Ошибка сохранения результата:", err)
				}
			}
		}
	}
}

func addProtocol(domain string) string {
	if !strings.HasPrefix(domain, "http") {
		return "http://" + domain
	}
	return domain
}
