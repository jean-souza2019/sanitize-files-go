package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var tmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>Processador de Arquivos</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f9;
            margin: 0;
            padding: 0;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
        }
        .container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            width: 300px;
            text-align: center;
        }
        h1 {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 10px;
            font-weight: bold;
        }
        input[type="text"] {
            width: calc(100% - 22px);
            padding: 10px;
            margin-bottom: 20px;
            border: 1px solid #ccc;
            border-radius: 4px;
        }
        button {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            background-color: #007bff;
            color: white;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover {
            background-color: #0056b3;
        }
        .result {
            margin-top: 20px;
        }
        .group-div {
            padding: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Processador de Arquivos</h1>
        <form id="fileForm" method="POST" action="/">
            <div class="group-div">
                <label for="directory">Caminho do Diretório:</label>
                <input type="text" id="directory" name="directory" required>
                <input type="file" id="filePicker" webkitdirectory directory style="display: none;">
                <button type="button" onclick="selectFolder()">Selecionar Pasta</button>
            </div>
            <button style="background-color: green;" type="submit">Processar Arquivos</button>
        </form>
        {{if .Started}}
        <div class="result">
            <h2>Resultados do Processamento</h2>
            <p>Início: {{.Started}}</p>
            <p>Fim: {{.Finished}}</p>
            <p>Duração: {{.Duration}} segundos</p>
            <p>Total de arquivos processados: {{.FileCount}}</p>
        </div>
        {{end}}
    </div>
    <script>
        function selectFolder() {
            document.getElementById('filePicker').click();
        }

        document.getElementById('filePicker').addEventListener('change', function(event) {
            const folderPath = event.target.files[0].webkitRelativePath.split('/')[0];
            document.getElementById('directory').value = folderPath;
        });
    </script>
</body>
</html>
`))

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

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	go openBrowser("http://localhost:8080")
	http.HandleFunc("/", handleProcess)
	log.Println("Acesse: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
