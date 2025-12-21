package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jomei/notionapi"
)

// --- CONFIGURAÇÕES LIDA DO AMBIENTE ---
var (
	notionToken = os.Getenv("NOTION_TOKEN")
	databaseID  = os.Getenv("DATABASE_ID")
)

func main() {
	if notionToken == "" || databaseID == "" {
		log.Fatal("❌ ERRO: Faltam as variáveis NOTION_TOKEN ou DATABASE_ID")
	}

	client := notionapi.NewClient(notionapi.Token(notionToken))

	// 1. FILTRO: Busca itens com a tag "Último Cap" (ajuste conforme seu Notion)
	query := &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Tags",
			Select: &notionapi.SelectFilterCondition{
				Equals: "Último Cap",
			},
		},
	}

	fmt.Println("🔍 Consultando Notion...")
	resp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseID), query)
	if err != nil {
		log.Fatalf("Erro na consulta: %v", err)
	}

	fmt.Printf("📚 Encontradas %d obras para verificar.\n", len(resp.Results))

	var wg sync.WaitGroup

	// 2. PARALELISMO (Goroutines)
	for _, page := range resp.Results {
		wg.Add(1)
		go func(p notionapi.Page) {
			defer wg.Done()
			checkManga(client, p)
		}(page)
	}

	wg.Wait()
	fmt.Println("🏁 Verificação concluída.")
}

func checkManga(client *notionapi.Client, page notionapi.Page) {
	props := page.Properties

	// Título (Safety Check)
	titleList, _ := props["Obra"].(*notionapi.TitleProperty)
	if len(titleList.Title) == 0 { return }
	title := titleList.Title[0].Text.Content

	// Link
	linkProp, ok := props["Link"].(*notionapi.URLProperty)
	if !ok || linkProp.URL == "" {
		fmt.Printf("⚠️ [%s] Sem URL cadastrada.\n", title)
		return
	}
	url := linkProp.URL

	// Capítulo Atual no Notion
	var currentCap float64
	if numProp, ok := props["Capítulo"].(*notionapi.NumberProperty); ok {
		currentCap = numProp.Number
	}

	// 3. SCRAPING
	latestCap, err := scrapeWeebCentral(url)
	if err != nil {
		fmt.Printf("❌ [%s] Erro no scrape: %v\n", title, err)
		return
	}

	// 4. LÓGICA DE ATUALIZAÇÃO
	if latestCap > currentCap {
		msg := fmt.Sprintf("🚀 [%s] Saiu o Cap %.1f (Você parou no %.1f)", title, latestCap, currentCap)
		fmt.Println(msg)
		
		// Atualiza o Notion
		updateNotion(client, page.ID, latestCap)
		
	} else {
		fmt.Printf("✅ [%s] Nada novo (%.1f)\n", title, latestCap)
	}
}

// Lógica de Scrape (Ajuste o seletor CSS se mudar o site)
func scrapeWeebCentral(url string) (float64, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	res, err := client.Do(req)
	if err != nil { return 0, err }
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("status http %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil { return 0, err }

	// SELETOR: Pega o primeiro link dentro da lista de capítulos
	// Ajuste: #chapter-list a (ou .text-lg, depende do site exato)
	text := doc.Find("#chapter-list a").First().Text()
	
	return extractNumber(text), nil
}

func updateNotion(client *notionapi.Client, pageID notionapi.PageID, newCap float64) {
	_, err := client.Page.Update(context.Background(), pageID, &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{
			"Capítulo": notionapi.NumberProperty{Number: newCap},
			// Opcional: Muda a tag para "Novo"
			"Tags": notionapi.SelectProperty{
				Select: notionapi.Option{Name: "Novo"},
			},
		},
	})
	if err != nil { fmt.Printf("Erro ao salvar no Notion: %v\n", err) }
}

func extractNumber(text string) float64 {
	re := regexp.MustCompile(`(\d+(\.\d+)?)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val
	}
	return 0
}