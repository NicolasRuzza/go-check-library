package httpex

import "time"

type Config struct {
	BaseTimeout time.Duration
	MaxRetries  int
	Exponent    int           // O multiplicador para o novo timeout de cada retry
	RetryWait   time.Duration // Tempo de espera entre as tentativas
}
