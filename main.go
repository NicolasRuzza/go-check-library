package main

import (
	"fmt"
	"go-check-library/notion"
	"go-check-library/scraper"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	token := os.Getenv("NOTION_TOKEN")
	dbId := os.Getenv("DATABASE_ID")

	if token == "" || dbId == "" {
		log.Fatal("Defina NOTION_TOKEN e DATABASE_ID")
	}

	notionClient := notion.NewClient(token, dbId)

	siteSelectors := map[string]string{
		"weebcentral.com": "#chapter-list > div:nth-child(1) > a > span:nth-child(2) > span:nth-child(1)",
		"fliptru.com":     "",
		"tapas.io":        "",
	}

	fmt.Println("Buscando no Notion...")

	// 3. Buscar Dados
	pages, err := notionClient.QueryBooks()
	if err != nil {
		log.Fatalf("Falha ao buscar mangás: %v", err)
	}

	fmt.Printf("Encontrados: %d obras para verificar.\n", len(pages))

	fmt.Print("\n xxxxxxxxx \n\n")

	for _, page := range pages {
		props := page.Properties
		url := props.Link.URL

		title := "Sem Título"
		if len(props.Obra.Title) > 0 {
			title = props.Obra.Title[0].PlainText
		}

		lastKnownChapter := props.UltimoCapConhecido.Number

		var selector, domain string

		for forDomain, forSelector := range siteSelectors {
			if strings.Contains(url, forDomain) {
				selector = forSelector
				domain = forDomain

				break
			}
		}

		if selector == "" {
			fmt.Printf("Site (%s) ainda não tratado\n", domain)
			continue
		}

		fmt.Printf(
			"Verificando obra [%s] no site [%s]. (Cap Corrente: %.0f | Ult cap conhecido: %.0f)\n",
			title, domain, props.Capitulo.Number, lastKnownChapter,
		)

		chapterFound, err := scraper.ScrapeLatestChapter(url, selector)
		if err != nil {
			fmt.Printf("Erro ao ler [%s]: %v\n", title, err)
			continue
		}

		if lastKnownChapter == 0 {
			fmt.Printf("Primeira sincronização para [%s]: %.1f\n", title, chapterFound)

			updateData := notion.UpdateProperties{
				UltimoCap: &notion.NumberProperty{
					Number: chapterFound,
				},
				// nil para não alterar a tag atual
				Tags: nil,
			}

			err := notionClient.UpdateChapter(page.ID, updateData)
			if err != nil {
				fmt.Printf("Erro ao salvar no Notion: %v\n", err)
			} else {
				fmt.Println("Notion atualizado com sucesso!")
			}
		} else if chapterFound > lastKnownChapter {
			fmt.Printf("NOVO CAPÍTULO ENCONTRADO! [%s] %.1f -> %.1f\n", title, lastKnownChapter, chapterFound)

			updateData := notion.UpdateProperties{
				UltimoCap: &notion.NumberProperty{
					Number: chapterFound,
				},
				Tags: &notion.SelectProperty{
					Select: notion.SelectOption{Name: "Novo Cap"},
				},
			}

			err := notionClient.UpdateChapter(page.ID, updateData)
			if err != nil {
				fmt.Printf("Erro ao salvar no Notion: %v\n", err)
			} else {
				fmt.Println("Notion atualizado com sucesso!")
			}
		} else {
			fmt.Printf("Nada novo. %s (Site: %.1f | Notion: %.1f)\n", title, chapterFound, lastKnownChapter)
		}

		if chapterFound > 0 {
			fmt.Printf(
				"[%s] (Cap Corrente: %.0f | Ult cap conhecido: %.0f)\n",
				title, props.Capitulo.Number, chapterFound,
			)
		}

		fmt.Print("\n --------- \n\n")

		time.Sleep(5 * time.Second)
	}
}
