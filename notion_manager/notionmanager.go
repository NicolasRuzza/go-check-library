package notionmanager

import (
	"encoding/json"
	"fmt"
	"go-check-library/httpex"
	"go-check-library/notion"
	"io"
	"net/http"
)

type NotionManager struct {
	Token string
	DbId  string
	Http  *httpex.HttpEx
}

func NewNotionManager(token, dbId string) *NotionManager {
	return &NotionManager{
		Token: token,
		DbId:  dbId,
		Http:  httpex.NewHttpEx(),
	}
}

func (nmg *NotionManager) QueryBooks() ([]notion.Page, error) {
	filter := notion.FilterBody{
		Filter: notion.AndFilter{
			And: []interface{}{
				notion.OrFilter{
					Or: []interface{}{
						notion.TagFilterSelect{
							Property: "Tags",
							Select: notion.Condition{
								Equals: "Último Cap",
							},
						},
						notion.TagFilterSelect{
							Property: "Tags",
							Select: notion.Condition{
								Equals: "Em Progresso",
							},
						},
						notion.TagFilterSelect{
							Property: "Tags",
							Select: notion.Condition{
								Equals: "Não Lança Cap",
							},
						},
						notion.TagFilterSelect{
							Property: "Tags",
							Select: notion.Condition{
								Equals: "Novo Cap",
							},
						},
						notion.TagFilterSelect{
							Property: "Tags",
							Select: notion.Condition{
								Equals: "Congelado",
							},
						},
					},
				},
				notion.NumberFilter{
					Property: "Capítulo",
					Number: notion.NumberCondition{
						GreaterThan: 0,
					},
				},
				notion.URLFilter{
					Property: "Link",
					URL: notion.URLCondition{
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

	var result notion.QueryResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

func (nmg *NotionManager) UpdateChapter(pageId string, newData notion.UpdateProperties) error {
	payload := notion.UpdatePageBody{
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

func (nmg *NotionManager) setNotionHeaders(request *http.Request) error {
	request.Header.Set("Authorization", "Bearer "+nmg.Token)
	request.Header.Set("Notion-Version", "2022-06-28")
	request.Header.Set("Content-Type", "application/json")

	return nil
}
