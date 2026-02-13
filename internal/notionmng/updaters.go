package notionmng

type UpdatePageBody struct {
	Properties UpdateProperties `json:"properties"`
}

type UpdateProperties struct {
	UltimoCap *NumberProperty `json:"Último Cap Conhecido,omitempty"`
	Tags      *SelectProperty `json:"Tags,omitempty"`
}
