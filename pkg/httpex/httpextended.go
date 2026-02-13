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
	config Config
}

func New(cfg Config) *HttpEx {
	if cfg.BaseTimeout == 0 {
		cfg.BaseTimeout = 10 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Exponent == 0 {
		cfg.Exponent = 2
	}
	if cfg.RetryWait == 0 {
		cfg.RetryWait = 5
	}

	return &HttpEx{config: cfg}
}

func CreateHttpRequest(method, url string, payload any) (*http.Request, error) {
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

	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return request, nil
}

func (httpex *HttpEx) DoWithRetry(request *http.Request) (*http.Response, error) {
	// Comeca com 10 segundos
	timeout := httpex.config.BaseTimeout * time.Second
	maxRetries := httpex.config.MaxRetries

	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		client := &http.Client{
			Timeout: timeout,
		}

		response, err := client.Do(request)

		// 2xx
		if err == nil && response.StatusCode >= 200 && response.StatusCode < 300 {
			return response, nil
		}

		lastErr = err

		msgErro := "Erro de conexao"
		if response != nil {
			// Limita leitura do erro a 4KB para nao estourar memoria com um html gigante
			limitReader := io.LimitReader(response.Body, 4096)
			body, _ := io.ReadAll(limitReader)
			response.Body.Close()

			msgErro = fmt.Sprintf("Status %d: %s", response.StatusCode, string(body))
		} else if err != nil {
			msgErro = err.Error()
		}

		fmt.Printf("ALERTA!!! Tentativa %d/%d falhou (%s). Timeout era %v.\n", i+1, maxRetries+1, msgErro, timeout)

		if i == maxRetries {
			break
		}

		timeout = timeout * time.Duration(httpex.config.Exponent)

		// Resetar o Body da requisicao
		// Como o Go le o body ao enviar, se enviar de novo sem resetar,
		// ele vai enviar um corpo vazio. O GetBody recria o leitor.
		if request.GetBody != nil && request.Body != nil {
			newBody, err := request.GetBody()

			if err != nil {
				return nil, fmt.Errorf("Não foi possível reconstruir o corpo da requisição para retry: %w", err)
			}

			request.Body = newBody
		}

		fmt.Printf("Tentando novamente com timeout de %v...\n", timeout)
		time.Sleep(httpex.config.RetryWait * time.Second)
	}

	return nil, fmt.Errorf("Falha apos todas as tentativas. Ultimo erro: %v", lastErr)
}
