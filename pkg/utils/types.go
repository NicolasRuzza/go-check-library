package utils

type TerminalANSIColor string

const (
	Reset  TerminalANSIColor = "\033[0m"
	Red    TerminalANSIColor = "\033[31m"
	Green  TerminalANSIColor = "\033[32m"
	Yellow TerminalANSIColor = "\033[33m"
	Blue   TerminalANSIColor = "\033[34m"
	Cyan   TerminalANSIColor = "\033[36m"
	Gray   TerminalANSIColor = "\033[37m"
)
