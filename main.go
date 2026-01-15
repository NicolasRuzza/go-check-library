package main

import (
	"fmt"
	"go-check-library/notion"
	notionmanager "go-check-library/notion_manager"
	"go-check-library/scraper"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	NumWorkers = 3
)

var siteSelectors = map[string]string{
	"weebcentral.com": "#chapter-list > div:nth-child(1) > a > span:nth-child(2) > span:nth-child(1)",
	"fliptru.com":     "",
	"tapas.io":        "",
}

func main() {
	token := os.Getenv("NOTION_TOKEN")
	dbId := os.Getenv("DATABASE_ID")

	if token == "" || dbId == "" {
		log.Fatal("Defina NOTION_TOKEN e DATABASE_ID")
	}

	nmg := notionmanager.NewNotionManager(token, dbId)

	fmt.Println("Buscando no Notion...")

	// 3. Buscar Dados
	pages, err := nmg.QueryBooks()
	if err != nil {
		log.Fatalf("Falha ao buscar mangás: %v", err)
	}

	totalBooks := len(pages)
	fmt.Printf("Encontrados: %d obras para verificar.\n", totalBooks)

	fmt.Print("\n xxxxxxxxx \n\n")

	jobs := make(chan notion.Page, totalBooks)
	results := make(chan string, totalBooks)

	var wg sync.WaitGroup

	for i := 1; i <= NumWorkers; i++ {
		wg.Add(1)
		go ScrapAndSave(nmg, i, jobs, results, &wg)
	}

	for _, page := range pages {
		jobs <- page
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	processedCount := 0
	for result := range results {
		processedCount++
		fmt.Println(result)
	}

	fmt.Println("Processamento finalizado!!!")
}

func ScrapAndSave(nmg *notionmanager.NotionManager, workerId int, jobs <-chan notion.Page, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for page := range jobs {
		props := page.Properties
		url := props.Link.URL

		title := "Sem Título"
		if len(props.Obra.Title) > 0 {
			title = props.Obra.Title[0].PlainText
		}

		lastKnownChapter := props.UltimoCapConhecido.Number

		selector, domain := getDomainSelector(url)
		if selector == "" {
			fmt.Printf("[Worker %d] Site (%s) ainda não tratado\n", workerId, domain)
			continue
		}

		fmt.Printf(
			"[Worker %d] Verificando obra [%s] no site [%s]. (Cap Corrente: %.0f | Ult cap conhecido: %.0f)\n",
			workerId, title, domain, props.Capitulo.Number, lastKnownChapter,
		)

		chapterFound, err := scraper.ScrapeLatestChapter(url, selector)
		if err != nil {
			fmt.Printf("[Worker %d] Erro ao ler [%s]: %v\n", workerId, title, err)
			continue
		}

		if lastKnownChapter == 0 {
			fmt.Printf("[Worker %d] Primeira sincronização para [%s]: %.1f\n", workerId, title, chapterFound)

			updateData := notion.UpdateProperties{
				UltimoCap: &notion.NumberProperty{
					Number: chapterFound,
				},
				// nil para não alterar a tag atual
				Tags: nil,
			}

			err := nmg.UpdateChapter(page.ID, updateData)
			if err != nil {
				fmt.Printf("[Worker %d] Erro ao salvar no Notion: %v\n", workerId, err)
			} else {
				fmt.Printf("[Worker %d] Notion atualizado com sucesso!\n", workerId)
			}
		} else if chapterFound > lastKnownChapter {
			fmt.Printf("[Worker %d] NOVO CAPÍTULO ENCONTRADO! [%s] %.1f -> %.1f\n", workerId, title, lastKnownChapter, chapterFound)

			updateData := notion.UpdateProperties{
				UltimoCap: &notion.NumberProperty{
					Number: chapterFound,
				},
				Tags: &notion.SelectProperty{
					Select: notion.SelectOption{Name: "Novo Cap"},
				},
			}

			err := nmg.UpdateChapter(page.ID, updateData)
			if err != nil {
				fmt.Printf("[Worker %d] Erro ao salvar no Notion: %v\n", workerId, err)
			} else {
				fmt.Printf("[Worker %d] Notion atualizado com sucesso!\n", workerId)
			}
		} else {
			fmt.Printf("[Worker %d] Nada novo. %s (Site: %.1f | Notion: %.1f)\n", workerId, title, chapterFound, lastKnownChapter)
		}

		if chapterFound > 0 {
			fmt.Printf(
				"[Worker %d] [%s] (Cap Corrente: %.0f | Ult cap conhecido: %.0f)\n",
				workerId, title, props.Capitulo.Number, chapterFound,
			)
		}

		fmt.Print("\n --------- \n\n")

		results <- fmt.Sprintf("[Worker %d] Processado: %s", workerId, title)

		// rand.Intn(3) retorna 0, 1 ou 2. Somando 3, temos 3, 4 ou 5.
		randomDelay := time.Duration(3+rand.Intn(3)) * time.Second
		time.Sleep(randomDelay)
	}
}

func getDomainSelector(url string) (string, string) {
	for domain, selector := range siteSelectors {
		if strings.Contains(url, domain) {
			return selector, domain
		}
	}

	return "", ""
}
