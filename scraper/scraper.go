package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func ScrapeLatestChapter(url string, cssSelector string) (float64, error) {
	var chapterText string

	collector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)
	if collector == nil {
		return 0, fmt.Errorf("falha ao criar o collector")
	}

	// Define o timeout para não ficar travado
	collector.SetRequestTimeout(30 * time.Second)

	// Callback: Quando encontrar o elemento
	collector.OnHTML(cssSelector, func(e *colly.HTMLElement) {
		// Pega o primeiro que encontrar e para (evita pegar lista inteira)
		if chapterText == "" {
			chapterText = strings.TrimSpace(e.Text)
		}
	})

	err := collector.Visit(url)
	if err != nil {
		return 0, fmt.Errorf("falha ao visitar %s: %v", url, err)
	}

	if chapterText == "" {
		return 0, fmt.Errorf("seletor '%s' nao encontrou nada na pagina", cssSelector)
	}

	return extractNumber(chapterText)
}

func extractNumber(text string) (float64, error) {
	// Regex para pegar numeros inteiros ou decimais
	re := regexp.MustCompile(`\d+([.,]\d+)?`)
	match := re.FindString(text)

	if match == "" {
		return 0, fmt.Errorf("nenhum numero encontrado no texto: %s", text)
	}

	// Troca virgula por ponto para converter corretamente
	match = strings.Replace(match, ",", ".", 1)

	return strconv.ParseFloat(match, 64)
}
