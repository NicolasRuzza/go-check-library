package notionmng

import (
	"encoding/json"
	"fmt"
	"go-check-library/httpex"
	"io"
	"net/http"
)

type NotionService struct {
	Token string
	DbId  string
	Http  *httpex.HttpEx
}

func New(token, dbId string) *NotionService {
	return &NotionService{
		Token: token,
		DbId:  dbId,
		Http:  httpex.NewHttpEx(),
	}
}

func (nmg *NotionService) QueryBooks() ([]Page, error) {
	// Tags para serem buscadas
	targetTags := []string{"Último Cap", "Em Progresso", "Não Lança Cap", "Novo Cap", "Congelado"}

	var orFilters []any
	for _, tag := range targetTags {
		orFilters = append(orFilters,
			TagFilterSelect{
				Property: "Tags",
				Select: Condition{
					Equals: tag,
				},
			},
		)
	}

	filter := FilterBody{
		Filter: AndFilter{
			And: []any{
				OrFilter{
					Or: orFilters,
				},
				NumberFilter{
					Property: "Capítulo",
					Number: NumberCondition{
						GreaterThan: 0,
					},
				},
				URLFilter{
					Property: "Link",
					URL: URLCondition{
						IsNotEmpty: true,
					},
				},
			},
		},
	}

	url := "https://api.notion.com/v1/databases/" + nmg.DbId + "/query"
	request, err := httpex.CreateHttpRequest("POST", url, filter)
	if err != nil {
		return nil, err
	}

	nmg.setNotionHeaders(request)

	response, err := nmg.Http.DoWithRetry(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("Erro API Notion (%d): %s", response.StatusCode, string(body))
	}

	var result QueryResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

func (nmg *NotionService) UpdateChapter(pageId string, newData UpdateProperties) error {
	payload := UpdatePageBody{
		Properties: newData,
	}

	url := "https://api.notion.com/v1/pages/" + pageId
	request, err := httpex.CreateHttpRequest("PATCH", url, payload)
	if err != nil {
		return err
	}

	nmg.setNotionHeaders(request)

	response, err := nmg.Http.DoWithRetry(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("Erro notion update (%d): %s", response.StatusCode, string(body))
	}

	return nil
}

func (nmg *NotionService) setNotionHeaders(request *http.Request) {
	request.Header.Set("Authorization", "Bearer "+nmg.Token)
	request.Header.Set("Notion-Version", "2022-06-28")
	request.Header.Set("Content-Type", "application/json")
}
