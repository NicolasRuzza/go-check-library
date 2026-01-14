package notion

// Estruturas de RESPOSTA (O que o Notion devolve)
type QueryResponse struct {
	Results []Page `json:"results"`
}

type Page struct {
	ID         string     `json:"id"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	Obra               TitleProperty  `json:"Obra"`
	Link               URLProperty    `json:"Link"`
	Capitulo           NumberProperty `json:"Capítulo"`
	UltimoCapConhecido NumberProperty `json:"Último Cap Conhecido"`
	Tipo               SelectProperty `json:"Tipo"`
	Tags               SelectProperty `json:"Tags"`
}
