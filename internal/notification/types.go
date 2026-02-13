package notification

// NtfyAction define os botões interativos na notificação
// No momento, vou limitar a implementacao somente ao view
type NtfyAction struct {
	Action NtfyActionType `json:"action"`          // "view", "broadcast", "http"
	Label  string         `json:"label"`           // O texto do botão
	Url    string         `json:"url,omitempty"`   // Para action="view"
	Clear  bool           `json:"clear,omitempty"` // Se true, limpa a notificação ao clicar
}

// Para mais detalhes, veja: https://docs.ntfy.sh/publish/?h=action#publish-as-json
type NtfyPayload struct {
	Topic   string `json:"topic"`
	Message string `json:"message,omitempty"`
	Title   string `json:"title,omitempty"` // Titulo = Tags + Title
	// Para mais Tags, veja: https://docs.ntfy.sh/emojis/
	// Tags sao emojis. Elas aparecem por escrito na parte debaixo da notificacao caso nao haja um emoji correspondente
	Tags     []string             `json:"tags,omitempty"`
	Priority NotificationPriority `json:"priority,omitempty"` // 1 (min) a 5 (max)

	// Ações e Interatividade
	Click   string       `json:"click,omitempty"`   // Evento 'on click' na notificacao para uma URL informada
	Actions []NtfyAction `json:"actions,omitempty"` // Botões de acao

	// Anexos e Ícones
	Attach   string `json:"attach,omitempty"`   // URL de uma imagem/arquivo
	Filename string `json:"filename,omitempty"` // Nome do arquivo do anexo
	Icon     string `json:"icon,omitempty"`     // URL do icone da notificacao

	// Formatação e Entrega
	Markdown bool   `json:"markdown,omitempty"` // Se true, processa negrito/italico na mensagem
	Delay    string `json:"delay,omitempty"`    // Ex: "30min", "9am". Veja: https://docs.ntfy.sh/publish/?h=actions#scheduled-delivery
	Email    string `json:"email,omitempty"`    // Enviar copia por email
}
