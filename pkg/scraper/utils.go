package scraper

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

func getRandomUserAgent() string {
	randomIndex := rand.Intn(len(userAgents))
	return userAgents[randomIndex]
}

func extractNumber(text string) (float64, error) {
	// Regex para pegar numeros inteiros ou decimais
	re := regexp.MustCompile(`\d+([.,]\d+)?`)
	match := re.FindString(text)

	if match == "" {
		return 0, fmt.Errorf("Nenhum numero encontrado no texto: %s", text)
	}

	// Troca virgula por ponto para converter corretamente
	match = strings.Replace(match, ",", ".", 1)

	return strconv.ParseFloat(match, 64)
}
