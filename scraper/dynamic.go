package scraper

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Eh headless
func ScrapeDynamicPage(url string, selector string) (float64, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(getRandomUserAgent()),
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("ignore-certificate-errors", true),
	)

	// Verifica se existe uma variavel de ambiente dizendo onde o Chrome esta.
	// No Dockerfile é definido CHROME_BIN=/usr/bin/chromium-browser
	chromeBin := os.Getenv("CHROME_BIN")
	if chromeBin != "" {
		opts = append(opts, chromedp.ExecPath(chromeBin))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Cria uma aba
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var textFound string

	err := chromedp.Run(ctx,
		// Vai para o site
		chromedp.Navigate(url),
		// Espera o seletor aparecer
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		// Pega o texto
		chromedp.Text(selector, &textFound, chromedp.ByQuery),
	)

	if err != nil {
		return 0, fmt.Errorf("Erro no browser: %v", err)
	}

	textFound = strings.TrimSpace(textFound)
	if textFound == "" {
		return 0, fmt.Errorf("Seletor '%s' retornou vazio", selector)
	}

	return extractNumber(textFound)
}
