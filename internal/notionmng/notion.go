package notionmng

import (
	"encoding/json"
	"fmt"
	"go-check-library/pkg/httpex"
	"net/http"
	"time"
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
		Http: httpex.New(httpex.Config{
			BaseTimeout: 10 * time.Second,
			MaxRetries:  3,
			Exponent:    2,
			RetryWait:   5 * time.Second,
		}),
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

	url := fmt.Sprintf("%s/databases/%s/query", NOTION_URL, nmg.DbId)
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

	url := fmt.Sprintf("%s/pages/%s", NOTION_URL, pageId)
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

	return nil
}

func (nmg *NotionService) setNotionHeaders(request *http.Request) {
	request.Header.Set("Authorization", "Bearer "+nmg.Token)
	request.Header.Set("Notion-Version", "2022-06-28")
}
