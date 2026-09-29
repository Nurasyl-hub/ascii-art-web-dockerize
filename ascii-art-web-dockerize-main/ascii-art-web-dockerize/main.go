package main

import (
	"bufio"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const artHeight = 8

type Glyph struct {
	Lines []string
}

type PageData struct {
	Text       string
	Banner     string
	Result     string
	ErrorMsg   string
	StatusCode int
}

var tpl *template.Template

var banners = map[string]string{
	"standard":   "banners/standard.txt",
	"shadow":     "banners/shadow.txt",
	"thinkertoy": "banners/thinkertoy.txt",
}

var bannerCache = map[string]map[rune]Glyph{}

func main() {
	var err error
	tpl, err = template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		log.Fatal("Ошибка при загрузке шаблона:", err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ascii-art", handleAsciiArt)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	renderPage(w, PageData{StatusCode: http.StatusOK})
}

func handleAsciiArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request: can't parse form", http.StatusBadRequest)
		return
	}

	text := strings.ReplaceAll(r.FormValue("text"), "\\n", "\n")
	banner := r.FormValue("banner")

	if text == "" {
		renderPage(w, PageData{
			ErrorMsg:   "Empty text",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	file, ok := banners[banner]
	if !ok {
		renderPage(w, PageData{
			ErrorMsg:   "Banner not found",
			StatusCode: http.StatusNotFound,
		})
		return
	}

	glyphs, err := loadBannerCached(banner, file)
	if err != nil {
		renderPage(w, PageData{
			ErrorMsg:   "Internal Server Error: can't load banner",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	result, err := renderASCII(text, glyphs)
	if err != nil {
		renderPage(w, PageData{
			ErrorMsg:   err.Error(),
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	renderPage(w, PageData{
		Text:       text,
		Banner:     banner,
		Result:     result,
		StatusCode: http.StatusOK,
	})
}

func renderPage(w http.ResponseWriter, data PageData) {
	w.WriteHeader(data.StatusCode)
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func renderASCII(input string, glyphs map[rune]Glyph) (string, error) {
	if !utf8.ValidString(input) {
		return "", errors.New("invalid UTF-8 input")
	}

	var out strings.Builder
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		if line == "" {
			out.WriteByte('\n')
			continue
		}
		for i := 0; i < artHeight; i++ {
			for _, ch := range line {
				glyph, ok := glyphs[ch]
				if !ok {
					continue
				}
				if i < len(glyph.Lines) {
					out.WriteString(glyph.Lines[i])
				}
			}
			out.WriteByte('\n')
		}
	}
	return out.String(), nil
}

func loadBannerCached(name, path string) (map[rune]Glyph, error) {
	if m, ok := bannerCache[name]; ok {
		return m, nil
	}
	m, err := loadBanner(path)
	if err != nil {
		return nil, err
	}
	bannerCache[name] = m
	return m, nil
}

func loadBanner(path string) (map[rune]Glyph, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	glyphs := make(map[rune]Glyph)
	r := rune(32)
	for i := 0; i+artHeight <= len(lines); i += artHeight + 1 {
		glyphs[r] = Glyph{Lines: lines[i : i+artHeight]}
		r++
	}
	return glyphs, nil
}
