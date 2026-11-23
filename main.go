package main

import (
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

//go:embed templates/*.html static/*.css static/*.js
var content embed.FS

func main() {
	addr := flag.String("addr", ":8080", "адрес для прослушивания")
	dbPath := flag.String("db", "menu.db", "путь к файлу SQLite")
	imgDir := flag.String("img", "static/img", "каталог с фотографиями блюд")
	flag.Parse()

	db, err := OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	tmpl := template.Must(template.ParseFS(content, "templates/*.html"))
	staticSub, err := fs.Sub(content, "static")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(*imgDir))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		menu, err := LoadMenu(db)
		if err != nil {
			log.Printf("load menu: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Restaurant": "Оттепель",
			"Tagline":    "Сезонное меню от шефа. Готовим из локальных продуктов каждый день.",
			"Menu":       menu,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			log.Printf("template: %v", err)
		}
	})

	log.Printf("Меню доступно на http://localhost%s (БД: %s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
