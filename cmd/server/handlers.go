package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"
)

var proverbs = []string{
	"Don't communicate by sharing memory, share memory by communicating.",
	"Concurrency is not parallelism.",
	"Channels orchestrate; mutexes serialize.",
	"The bigger the interface, the weaker the abstraction.",
	"Make the zero value useful.",
	"interface{} says nothing.",
	"Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.",
	"A little copying is better than a little dependency.",
	"Clear is better than clever.",
	"Reflection is never clear.",
	"Errors are values.",
	"Don't just check errors, handle them gracefully.",
	"Design the architecture, name the components, document the details.",
	"Documentation is for users.",
	"Don't panic.",
}

type PageData struct {
	Title       string
	Description string
	Year        int
	Version     string
}

type templateCache map[string]*template.Template

func newTemplateCache(templatesFS fs.FS) (templateCache, error) {
	cache := make(templateCache)

	pages := []string{
		"pages/home.html",
		"pages/404.html",
	}

	for _, page := range pages {
		patterns := []string{
			"base.html",
			"partials/*.html",
			page,
		}

		tmpl, err := template.New("base").ParseFS(templatesFS, patterns...)
		if err != nil {
			return nil, fmt.Errorf("parsing template %q: %w", page, err)
		}

		cache[page] = tmpl
	}

	return cache, nil
}

func render(w http.ResponseWriter, logger *slog.Logger, status int, tmpl *template.Template, data PageData) {
	buf := new(bytes.Buffer)

	if err := tmpl.ExecuteTemplate(buf, "base", data); err != nil {
		logger.Error("template execution failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}
func homeHandler(cache templateCache, logger *slog.Logger, version string) http.HandlerFunc {
	tmpl, exists := cache["pages/home.html"]
	if !exists {
		panic("pages/home.html not found in template cache")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:       "Home",
			Description: "DAVID - Backend developer focused on Go, clean architecture, systems engineering, and high-performance APIs.",
			Year:        time.Now().UTC().Year(),
			Version:     version,
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")
		render(w, logger, http.StatusOK, tmpl, data)
	}
}

func notFoundHandler(cache templateCache, logger *slog.Logger, version string) http.HandlerFunc {
	tmpl, exists := cache["pages/404.html"]
	if !exists {
		panic("pages/404.html not found in template cache")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:       "404 Not Found",
			Description: "The page you are looking for does not exist.",
			Year:        time.Now().UTC().Year(),
			Version:     version,
		}

		render(w, logger, http.StatusNotFound, tmpl, data)
	}
}

func healthHandler(version string, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"status":     "healthy",
			"version":    version,
			"uptime_sec": int(time.Since(startTime).Seconds()),
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func gopherHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]string{
			"proverb": proverbs[rand.IntN(len(proverbs))],
			"source":  "https://go.dev/wiki/GoProverbs",
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.Error("failed to write gopher response", "error", err)
		}
	}
}
