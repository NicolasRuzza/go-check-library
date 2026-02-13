package notification

type NotificationPriority int

// Para mais detalhes, veja: https://docs.ntfy.sh/publish/#message-priority
const (
	MAX     NotificationPriority = 5 // Really long vibration bursts, default notification sound with a pop-over notification.
	HIGH    NotificationPriority = 4 // Long vibration burst, default notification sound with a pop-over notification.
	DEFAULT NotificationPriority = 3 // Short default vibration and sound. Default notification behavior.
	LOW     NotificationPriority = 2 // No vibration or sound. Notification will not visibly show up until notification drawer is pulled down.
	MIN     NotificationPriority = 1 // No vibration or sound. The notification will be under the fold in "Other notifications".
)

// Para mais detalhes, veja: https://docs.ntfy.sh/publish/?h=action#action-buttons
type NtfyActionType string

const (
	VIEW      NtfyActionType = "view"
	BROADCAST NtfyActionType = "broadcast"
	HTTP      NtfyActionType = "http"
)

type NotificationType int

const (
	NONE  NotificationType = 0
	NEW   NotificationType = 1
	FIRST NotificationType = 2
	FIX   NotificationType = 3
)

const (
	NOTIFICATION_URL = "https://ntfy.sh/"
	// O ntfy so consegue usar uma url como icone, logo eu fiz um commit so da imagem e agora estou usando-a
	PROJECT_ICON_URL = "https://github.com/NicolasRuzza/go-check-library/blob/feature/notification/assets/icon.png?raw=true"
)
