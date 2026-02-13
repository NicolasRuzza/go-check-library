package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationHelpers(t *testing.T) {
	// Server para receber a requisicao em vez de enviar direto para o ntfy e me notificar a cada notificacao testada
	var lastPayload NtfyPayload

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		json.NewDecoder(request.Body).Decode(&lastPayload)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := New("topico-teste")
	service.BaseURL = server.URL // Injeta a URL do server que recebera as notificacoes

	// Casos de teste
	scenarios := []struct {
		name             string
		run              func()
		expectedTitle    string
		expectedPriority NotificationPriority
		expectedTag      string
	}{
		{
			name:             "Erro de Conexão no Scraper",
			run:              func() { service.NotifyError(fmt.Errorf("timeout"), "Scraper Estático") },
			expectedTitle:    "ERRO CRÍTICO",
			expectedPriority: HIGH,
			expectedTag:      "warning",
		},
		{
			name:             "Informativo de Sincronização",
			run:              func() { service.NotifyInfo("Sincronismo", "Primeira carga concluída") },
			expectedTitle:    "Sincronismo",
			expectedPriority: LOW,
			expectedTag:      "information_source",
		},
		{
			name:             "Aviso de Nova Publicação",
			run:              func() { service.NotifyNewChapter("A Metamorfose", 2.0, "http://leitura.com") },
			expectedTitle:    "A Metamorfose - Novo Capítulo!",
			expectedPriority: DEFAULT,
			expectedTag:      "loud_sound",
		},
	}

	// Disparando
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Resetando o payload para cada teste
			lastPayload = NtfyPayload{}
			// Executa a funcao de cada teste (propriedade 'run') e verifica se o retorno bateu com o resultado esperado
			scenario.run()

			if lastPayload.Title != scenario.expectedTitle {
				t.Errorf("[%s] Título errado: %s", scenario.name, lastPayload.Title)
			}

			if lastPayload.Priority != scenario.expectedPriority {
				t.Errorf("[%s] Prioridade errada: %d", scenario.name, lastPayload.Priority)
			}

			if lastPayload.Icon == "" {
				t.Errorf("[%s] O ícone da notificação não foi enviado", scenario.name)
			}
		})
	}
}
