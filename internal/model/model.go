package model

type ChatRequest struct {
	Prompt   string
	Model    string
	Provider string
}

type ChatResponse struct {
	Content string
	Usage   int
	Tokens  int
	Model   string
}
