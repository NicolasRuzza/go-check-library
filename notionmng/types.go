package notionmng

// Helpers
type TitleProperty struct {
	Title []TextContent `json:"title"`
}
type URLProperty struct {
	URL string `json:"url"`
}
type NumberProperty struct {
	Number float64 `json:"number"`
}
type SelectProperty struct {
	Select SelectOption `json:"select"`
}
type SelectOption struct {
	Name string `json:"name"`
}
type TextContent struct {
	PlainText string `json:"plain_text"`
}
