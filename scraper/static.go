package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func ScrapeStaticPage(url string, cssSelector string) (float64, error) {
	var chapterText string

	collector := colly.NewCollector()
	if collector == nil {
		return 0, fmt.Errorf("Falha ao criar o coletor")
	}

	// Define o timeout para não ficar travado
	collector.SetRequestTimeout(30 * time.Second)

	// Antes de cada requisição, sorteia um User-Agent
	collector.OnRequest(func(r *colly.Request) {
		userAgent := getRandomUserAgent()

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
