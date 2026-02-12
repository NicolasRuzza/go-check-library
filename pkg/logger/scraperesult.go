package logger

import (
	"fmt"
	"go-check-library/pkg/utils"
)

type ScrapeResult struct {
	WorkerId int
	Domain   string
	Title    string
	Type     LogType
	Message  string
}

func (sr ScrapeResult) String() string {
	return fmt.Sprintf("[Worker %d] [%s] [%s] {%s} - %s",
		sr.WorkerId, sr.Domain, sr.Title, sr.Type, sr.Message)
}

func (sr ScrapeResult) ColoredString() string {
	baseMsg := fmt.Sprintf("[Worker %d] [%s] %s", sr.WorkerId, sr.Title, sr.Message)

	switch sr.Type {
	case SUCCESS:
		return fmt.Sprintf("%s %s%s", utils.Green, baseMsg, utils.Reset)
	case ERROR:
		return fmt.Sprintf("%s %s%s", utils.Red, baseMsg, utils.Reset)
	case WARN:
		return fmt.Sprintf("%s %s%s", utils.Yellow, baseMsg, utils.Reset)
	case INFO:
		return fmt.Sprintf("%s %s%s", utils.Cyan, baseMsg, utils.Reset)
	default:
		return baseMsg
	}
}
