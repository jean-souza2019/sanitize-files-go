package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var itemsCount = 0

func processFile(wg *sync.WaitGroup, filePath string, newBasePath string, _ time.Time) {
	defer wg.Done()
	fileName := filepath.Base(filePath)
	newFileName := sanitizeFileName(fileName)
	newFilePath := filepath.Join(newBasePath, newFileName)

	err := os.Rename(filePath, newFilePath)
	if err != nil {
		log.Printf("Erro ao renomear arquivo %s para %s: %v\n", filePath, newFilePath, err)
		return
	}
	itemsCount++
}

func sanitizeFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	cleanName := reg.ReplaceAllString(name, "")
	return strings.ToUpper(cleanName) + ext
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <inputDir>", os.Args[0])
	}

	dirPath := os.Args[1]

	startTime := time.Now()
	log.Printf("Iniciado processo em: %v\n", startTime)

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("Erro ao ler diretório: %v\n", err)
		return
	}

	var wg sync.WaitGroup

	for _, file := range files {
		if !file.IsDir() {
			wg.Add(1)
			go processFile(&wg, filepath.Join(dirPath, file.Name()), dirPath, startTime)
		}
	}

	wg.Wait()

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("Finalizado processo em: %v\n", endTime)
	log.Printf("Tempo duração: %v\n", duration)
	log.Printf("Arquivos processados: %v\n", itemsCount)
}
