package embedding

// type EmbeddingReq struct {
// 	Model string
// 	Prompt string
// }

type EmbeddingRes struct {
	Embedding []float32 `json:"embedding"`
}

type IEmbedding interface {
	GetEmbedding(model, Prompt string) (EmbeddingRes, error)
}