package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

func processFile(wg *sync.WaitGroup, filePath string, newBasePath string, counter *int) {
	defer wg.Done()
	fileName := filepath.Base(filePath)
	newFileName := sanitizeFileName(fileName)
	newFilePath := filepath.Join(newBasePath, newFileName)

	err := os.Rename(filePath, newFilePath)
	if err != nil {
		log.Printf("Erro ao renomear o arquivo %s para %s: %v\n", filePath, newFilePath, err)
		return
	}
	*counter++
}

func sanitizeFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	cleanName := reg.ReplaceAllString(name, "")
	return strings.ToUpper(cleanName) + ext
}

func formatTime(t time.Time) string {
	return t.Format("02/01/2006 15:04:05")
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	dirPath := r.FormValue("directory")
	startTime := time.Now()
	log.Printf("Iniciado o processamento em %v\n", formatTime(startTime))

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("Erro ao ler o diretório: %v\n", err)
		http.Error(w, "Erro ao ler o diretório", http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	counter := 0

	for _, file := range files {
		if !file.IsDir() {
			wg.Add(1)
			go processFile(&wg, filepath.Join(dirPath, file.Name()), dirPath, &counter)
		}
	}

	wg.Wait()

	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()
	log.Printf("Finalizado o processamento em %v\n", formatTime(endTime))
	log.Printf("Duração do processamento: %.2f segundos\n", duration)

	tmpl.Execute(w, struct {
		Started   string
		Finished  string
		Duration  float64
		FileCount int
	}{
		Started:   formatTime(startTime),
		Finished:  formatTime(endTime),
		Duration:  duration,
		FileCount: counter,
	})
}

func main() {
	http.HandleFunc("/", handleProcess)
	log.Println("Acesse: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
