package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NomicEmbedding struct {
	baseUrl string
	port int
	url string
}

func NewNomicEmbedding(baseUrl, url string) IEmbedding {
	return &NomicEmbedding{
		baseUrl: baseUrl,
		url: url,
	}
}

func (n *NomicEmbedding) GetEmbedding(model, prompt string) (EmbeddingRes, error) {
	// fmt.Println("NomicEmbedding : Calling Embedding Api :: ")
	uri := n.baseUrl + n.url
	// fmt.Println("NomicEmbedding : ", uri)

	body, err := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
	})
	if err != nil {
		fmt.Println("NomicEmbedding : Got error in body marshal ", err)
		return EmbeddingRes{}, err
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewReader(body))
	if err != nil {
		fmt.Println("NomicEmbedding : Got error in creating request ", err)
		return EmbeddingRes{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("NomicEmbedding : Got error in sending request ", err)
		return EmbeddingRes{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return EmbeddingRes{}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var resData EmbeddingRes
	if err := json.NewDecoder(resp.Body).Decode(&resData); err != nil {
		fmt.Println("NomicEmbedding : Got error in decoding response ", err)
		return EmbeddingRes{}, err
	}

	return resData, nil
}
