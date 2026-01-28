package httpex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpEx struct {
}

func NewHttpEx() *HttpEx {
	return &HttpEx{}
}

func CreateHttpRequest(method, url string, payload interface{}) (*http.Request, error) {
	var body io.Reader

	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("Erro ao criar json: %v", err)
		}

		body = bytes.NewBuffer(jsonBytes)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (httpex *HttpEx) DoWithRetry(request *http.Request) (*http.Response, error) {
	// Começa com 10 segundos
	timeout := 10 * time.Second
	maxRetries := 3

	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		client := &http.Client{
			Timeout: timeout,
		}

		response, err := client.Do(request)

		if err == nil && response.StatusCode == 200 {
			return response, nil
		}

		lastErr = err

		msgErro := "Erro de conexao"
		if response != nil {
			body, _ := io.ReadAll(response.Body)
			response.Body.Close()

			msgErro = fmt.Sprintf("Status %d: %s", response.StatusCode, string(body))
		} else if err != nil {
			msgErro = err.Error()
		}

		fmt.Printf("ALERTA!!! Tentativa %d/%d falhou (%s). Timeout era %v.\n", i+1, maxRetries+1, msgErro, timeout)

		if i == maxRetries {
			break
		}

		timeout = timeout * 5

		// Resetar o Body da requisicao
		// Como o Go le o body ao enviar, se enviar de novo sem resetar,
		// ele vai enviar um corpo vazio. O GetBody recria o leitor.
		if request.GetBody != nil {
			newBody, _ := request.GetBody()
			request.Body = newBody
		}

		fmt.Printf("Tentando novamente com timeout de %v...\n", timeout)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("Falha apos todas as tentativas. Ultimo erro: %v", lastErr)
}
