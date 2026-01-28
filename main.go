package main

import (
	"fmt"
	"go-check-library/logger"
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

const NumWorkers = 3

type SiteConfig struct {
	Selector   string
	UseBrowser bool // true -> conteudo dinamico e requer browser. false -> site estatico
}

var siteConfigs = map[string]SiteConfig{
	"weebcentral.com": {
		Selector:   "#chapter-list > div:nth-child(1) > a > span:nth-child(2) > span:nth-child(1)",
		UseBrowser: false,
	},
	"fliptru.com": {
		Selector:   ".chapter-list-item:nth-child(1) a",
		UseBrowser: true,
	},
	"tapas.io": {
		Selector:   "p.episode-cnt",
		UseBrowser: false,
	},
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

	jobs := make(chan notion.Page, totalBooks)
	results := make(chan logger.ScrapeResult, totalBooks)

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

	var allResults []logger.ScrapeResult
	var errors []logger.ScrapeResult
	var updates []logger.ScrapeResult

	for result := range results {
		allResults = append(allResults, result)

		// Ao vivo
		fmt.Println(result.ColoredString())

		if result.Type == logger.ERROR {
			errors = append(errors, result)
		}
		if result.Type == logger.SUCCESS {
			updates = append(updates, result)
		}
	}

	fmt.Println("\n\n================ RELATÓRIO FINAL ================")

	fmt.Printf("Total Processado: %d\n", len(allResults))
	fmt.Printf("Sucessos/Atualizações: %d\n", len(updates))
	fmt.Printf("Erros: %d\n", len(errors))

	if len(updates) > 0 {
		fmt.Println("\n--- OBRAS ATUALIZADAS ---")

		for _, up := range updates {
			fmt.Printf("-> %s: %s\n", up.Title, up.Message)
		}
	}

	if len(errors) > 0 {
		fmt.Println("\n--- FALHAS ---")

		for _, err := range errors {
			fmt.Printf("-> [%s] %s: %s\n", err.Domain, err.Title, err.Message)
		}
	}
}

func ScrapAndSave(nmg *notionmanager.NotionManager, workerId int, jobs <-chan notion.Page, results chan<- logger.ScrapeResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for page := range jobs {
		props := page.Properties
		url := props.Link.URL

		title := "Sem Título"
		if len(props.Obra.Title) > 0 {
			title = props.Obra.Title[0].PlainText
		}

		lastKnownChapter := props.UltimoCapConhecido.Number

		config, domain := getDomainConfig(url)

		resultBase := logger.ScrapeResult{
			WorkerId: workerId,
			Domain:   domain,
			Title:    title,
		}

		if config.Selector == "" {
			resultBase.Type = logger.WARN
			resultBase.Message = "Site não mapeado/tratado"
			results <- resultBase
			continue
		}

		var chapterFound float64
		var err error

		if config.UseBrowser {
			chapterFound, err = scraper.ScrapeDynamicPage(url, config.Selector)
		} else {
			chapterFound, err = scraper.ScrapeStaticPage(url, config.Selector)
		}

		if err != nil {
			resultBase.Type = logger.ERROR
			resultBase.Message = fmt.Sprintf("Falha no Scraper: %v", err)
			results <- resultBase
			continue
		}

		if chapterFound > lastKnownChapter && chapterFound > props.Capitulo.Number {
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
				resultBase.Type = logger.ERROR
				resultBase.Message = fmt.Sprintf("Erro ao atualizar novo: %v", err)
			} else {
				resultBase.Type = logger.SUCCESS
				resultBase.Message = fmt.Sprintf("Atualizado! %.1f -> %.1f", lastKnownChapter, chapterFound)
			}

			results <- resultBase
		} else if lastKnownChapter == 0 {
			updateData := notion.UpdateProperties{
				UltimoCap: &notion.NumberProperty{
					Number: chapterFound,
				},
				// nil para não alterar a tag atual
				Tags: nil,
			}

			err := nmg.UpdateChapter(page.ID, updateData)
			if err != nil {
				resultBase.Type = logger.ERROR
				resultBase.Message = fmt.Sprintf("Erro ao salvar (1ª carga): %v", err)
			} else {
				resultBase.Type = logger.SUCCESS
				resultBase.Message = fmt.Sprintf("Primeira sincronização: Cap %.1f", chapterFound)
			}

			results <- resultBase
		} else {
			if chapterFound != lastKnownChapter {
				updateData := notion.UpdateProperties{
					UltimoCap: &notion.NumberProperty{
						Number: chapterFound,
					},
					// nil para não alterar a tag atual
					Tags: nil,
				}

				err := nmg.UpdateChapter(page.ID, updateData)
				if err != nil {
					resultBase.Type = logger.ERROR
					resultBase.Message = fmt.Sprintf("Erro ao salvar (1ª carga): %v", err)
					results <- resultBase
					continue
				}
			}

			resultBase.Type = logger.INFO
			resultBase.Message = fmt.Sprintf("Nada novo. (Site: %.1f | Notion: %.1f)", chapterFound, lastKnownChapter)
			results <- resultBase
		}

		// rand.Intn(3) retorna 0, 1 ou 2. Somando 3, temos 3, 4 ou 5.
		randomDelay := time.Duration(3+rand.Intn(3)) * time.Second
		time.Sleep(randomDelay)
	}
}

func getDomainConfig(url string) (SiteConfig, string) {
	for domain, config := range siteConfigs {
		if strings.Contains(url, domain) {
			return config, domain
		}
	}

	return SiteConfig{}, ""
}
