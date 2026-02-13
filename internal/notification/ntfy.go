package notification

import (
	"fmt"
	"go-check-library/pkg/httpex"
	"time"
)

// Limitacoes do Ntfy.sh: https://docs.ntfy.sh/publish/?h=json#limitations

type NotificationService struct {
	// Topic eh o nome do canal de mensagens criado no ntfy.sh
	Topic   string
	Http    *httpex.HttpEx
	Icon    string
	BaseURL string
}

func New(topic string) *NotificationService {
	if topic == "" {
		panic("ERRO CRÍTICO: Tópico do Ntfy não configurado!")
	}

	return &NotificationService{
		Topic: topic,
		Http: httpex.New(httpex.Config{
			BaseTimeout: 5 * time.Second,
			MaxRetries:  2,
			Exponent:    1,
			RetryWait:   1 * time.Second,
		}),
		Icon:    PROJECT_ICON_URL,
		BaseURL: NOTIFICATION_URL,
	}
}

// Envia uma notificação para o seu celular via ntfy.sh
func (ntfn *NotificationService) sendPayload(payload NtfyPayload) {
	request, err := httpex.CreateHttpRequest("POST", ntfn.BaseURL, payload)
	if err != nil {
		panic(fmt.Errorf("Erro fatal ao criar request ntfy: %w", err))
	}

	response, err := ntfn.Http.DoWithRetry(request)
	if err != nil {
		panic(fmt.Errorf("Erro fatal de rede ntfy: %w", err))
	}

	defer response.Body.Close()
}

func (ntfn *NotificationService) Send(title string, message string, tags []string, priority NotificationPriority, actions []NtfyAction, markdown bool) {
	payload := NtfyPayload{
		Topic:    ntfn.Topic,
		Message:  message,
		Title:    title,
		Tags:     tags,
		Priority: priority,
		Actions:  actions,
		Markdown: markdown,
		Icon:     ntfn.Icon,
	}

	ntfn.sendPayload(payload)
}

func (ntfn *NotificationService) NotifyError(err error, context string) {
	tags := []string{"scroll", "warning"}

	msg := fmt.Sprintf("Falha em: **%s**\nErro: `%v`", context, err)

	ntfn.Send("ERRO CRÍTICO", msg, tags, HIGH, nil, true)
}

func (ntfn *NotificationService) NotifyInfo(title string, message string) {
	ntfn.Send(
		title,
		message,
		[]string{"scroll", "information_source"},
		LOW,
		nil,
		true,
	)
}

func (ntfn *NotificationService) NotifyNewChapter(title string, chapter float64, url string) {
	ntfnTitle := fmt.Sprintf("%s - Novo Capítulo!", title)

	msg := fmt.Sprintf("Capítulo **%.1f** já está disponível.", chapter)

	actions := []NtfyAction{
		{
			Action: "view",
			Label:  "Ler Agora 🚀",
			Url:    url,
			Clear:  true,
		},
	}

	ntfn.Send(
		ntfnTitle,
		msg,
		[]string{"scroll", "loud_sound"},
		DEFAULT,
		actions,
		true,
	)
}
