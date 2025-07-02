package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Result struct {
	Domain         string
	Category       string
	Word           string
	Context        string
	LLMFlag        bool
	LLMExplain     string
	ScreenshotPath string
	CreatedAt      string
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "results.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", resultsHandler)
	http.Handle("/screenshots/", http.StripPrefix("/screenshots/", http.FileServer(http.Dir("screenshots"))))

	fmt.Println("🌐 Веб-сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var rows *sql.Rows
	var err error
	if query == "" {
		rows, err = db.Query(`SELECT domain, category, word, context, llm_flag, llm_explain, screenshot_path, created_at FROM results ORDER BY created_at DESC LIMIT 100`)
	} else {
		rows, err = db.Query(`SELECT domain, category, word, context, llm_flag, llm_explain, screenshot_path, created_at FROM results WHERE domain LIKE ? OR category LIKE ? ORDER BY created_at DESC LIMIT 100`, "%"+query+"%", "%"+query+"%")
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	results := []Result{}
	for rows.Next() {
		var r Result
		err = rows.Scan(&r.Domain, &r.Category, &r.Word, &r.Context, &r.LLMFlag, &r.LLMExplain, &r.ScreenshotPath, &r.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		results = append(results, r)
	}

	tmpl := `
	<html>
	<head><title>Результаты модерации</title></head>
	<body>
		<h1>Результаты модерации</h1>
		<form method="GET">
			<input name="q" placeholder="Поиск по домену или категории" value="{{.Query}}">
			<button type="submit">Поиск</button>
		</form>
		<table border="1" cellpadding="5" cellspacing="0">
			<tr>
				<th>Домен</th><th>Категория</th><th>Найденное слово</th><th>Контекст</th><th>LLM флаг</th><th>Объяснение</th><th>Скриншот</th><th>Дата</th>
			</tr>
			{{range .Results}}
			<tr>
				<td>{{.Domain}}</td>
				<td>{{.Category}}</td>
				<td>{{.Word}}</td>
				<td><pre style="max-width:400px;white-space: pre-wrap;">{{.Context}}</pre></td>
				<td>{{.LLMFlag}}</td>
				<td>{{.LLMExplain}}</td>
				<td>{{if .ScreenshotPath}}<a href="/screenshots/{{.ScreenshotPath | base}}">Скриншот</a>{{end}}</td>
				<td>{{.CreatedAt}}</td>
			</tr>
			{{end}}
		</table>
	</body>
	</html>`

	funcMap := template.FuncMap{
		"base": func(path string) string {
			parts := strings.Split(path, string(os.PathSeparator))
			return parts[len(parts)-1]
		},
	}

	t, err := template.New("results").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := struct {
		Results []Result
		Query   string
	}{
		Results: results,
		Query:   query,
	}

	t.Execute(w, data)
}
