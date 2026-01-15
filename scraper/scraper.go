package scraper

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

func ScrapeLatestChapter(url string, cssSelector string) (float64, error) {
	var chapterText string

	collector := colly.NewCollector()
	if collector == nil {
		return 0, fmt.Errorf("Falha ao criar o coletor")
	}

	// Define o timeout para não ficar travado
	collector.SetRequestTimeout(30 * time.Second)

	// Antes de cada requisição, sorteia um User-Agent
	collector.OnRequest(func(r *colly.Request) {
		// Pega um índice aleatório
		randomIndex := rand.Intn(len(userAgents))
		userAgent := userAgents[randomIndex]

		// Seta no header
		r.Headers.Set("User-Agent", userAgent)

		r.Headers.Set("Referer", "https://google.com")
	})

	// Callback: Quando encontrar o elemento
	collector.OnHTML(cssSelector, func(e *colly.HTMLElement) {
		// Pega o primeiro que encontrar e para (evita pegar lista inteira)
		if chapterText == "" {
			chapterText = strings.TrimSpace(e.Text)
		}
	})

	err := collector.Visit(url)
	if err != nil {
		return 0, fmt.Errorf("Falha ao visitar %s: %v", url, err)
	}

	if chapterText == "" {
		return 0, fmt.Errorf("Seletor '%s' nao encontrou nada na pagina", cssSelector)
	}

	return extractNumber(chapterText)
}

func extractNumber(text string) (float64, error) {
	// Regex para pegar numeros inteiros ou decimais
	re := regexp.MustCompile(`\d+([.,]\d+)?`)
	match := re.FindString(text)

	if match == "" {
		return 0, fmt.Errorf("Nenhum numero encontrado no texto: %s", text)
	}

	// Troca virgula por ponto para converter corretamente
	match = strings.Replace(match, ",", ".", 1)

	return strconv.ParseFloat(match, 64)
}
