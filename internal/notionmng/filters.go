package notionmng

type FilterBody struct {
	Filter any `json:"filter"`
}

// Aceita qualquer tipo de filtro (Select, MultiSelect, etc)
type AndFilter struct {
	And []any `json:"and"`
}
type OrFilter struct {
	Or []any `json:"or"`
}

// Estrutura para colunas tipo "Select" (Caixa unica)
type TagFilterSelect struct {
	Property string    `json:"property"`
	Select   Condition `json:"select"`
}

// Estrutura para colunas tipo "Multi-select" (Varias tags)
type TagFilterMultiSelect struct {
	Property    string    `json:"property"`
	MultiSelect Condition `json:"multi_select"`
}

type Condition struct {
	Equals   string `json:"equals,omitempty"`
	Contains string `json:"contains,omitempty"`
}

type NumberFilter struct {
	Property string          `json:"property"`
	Number   NumberCondition `json:"number"`
}
type NumberCondition struct {
	GreaterThan float64 `json:"greater_than"` // > 0
	// Poderia ter LessThan, Equals, etc.
}

type URLFilter struct {
	Property string       `json:"property"`
	URL      URLCondition `json:"url"`
}
type URLCondition struct {
	IsNotEmpty bool   `json:"is_not_empty,omitempty"`
	IsEmpty    bool   `json:"is_empty,omitempty"`
	Equals     string `json:"equals,omitempty"`
	Contains   string `json:"contains,omitempty"`
}
