package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	Token string
	DbId  string
	Http  *http.Client
}

func NewClient(token, dbId string) *Client {
	return &Client{
		Token: token,
		DbId:  dbId,
		Http:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (client *Client) QueryBooks() ([]Page, error) {
	filter := FilterBody{
		Filter: AndFilter{
			And: []interface{}{
				OrFilter{
					Or: []interface{}{
						TagFilterSelect{
							Property: "Tags",
							Select: Condition{
								Equals: "Último Cap",
							},
						},
						TagFilterSelect{
							Property: "Tags",
							Select: Condition{
								Equals: "Em Progresso",
							},
						},
						TagFilterSelect{
							Property: "Tags",
							Select: Condition{
								Equals: "Não Lança Cap",
							},
						},
						TagFilterSelect{
							Property: "Tags",
							Select: Condition{
								Equals: "Novo Cap",
							},
						},
						TagFilterSelect{
							Property: "Tags",
							Select: Condition{
								Equals: "Congelado",
							},
						},
					},
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

	url := "https://api.notion.com/v1/databases/" + client.DbId + "/query"
	request, err := client.createHttpRequest("POST", url, filter)
	if err != nil {
		return nil, err
	}

	response, err := client.Http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("erro API Notion (%d): %s", response.StatusCode, string(body))
	}

	var result QueryResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

func (client *Client) UpdateChapter(pageId string, newData UpdateProperties) error {
	payload := UpdatePageBody{
		Properties: newData,
	}

	url := "https://api.notion.com/v1/pages/" + pageId
	request, err := client.createHttpRequest("PATCH", url, payload)
	if err != nil {
		return err
	}

	response, err := client.Http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("erro notion update (%d): %s", response.StatusCode, string(body))
	}

	return nil
}

func (client *Client) setNotionHeaders(request *http.Request) error {
	request.Header.Set("Authorization", "Bearer "+client.Token)
	request.Header.Set("Notion-Version", "2022-06-28")
	request.Header.Set("Content-Type", "application/json")

	return nil
}

func (client *Client) createHttpRequest(method, url string, payload interface{}) (*http.Request, error) {
	var body io.Reader

	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar json: %v", err)
		}

		body = bytes.NewBuffer(jsonBytes)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	client.setNotionHeaders(request)

	return request, nil
}
