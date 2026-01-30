package embedding

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type EmbeddingRequest struct {
	Texts []string `json:"texts"`
}

type EmbeddingResponseValue struct {
	Embeddings [][]float64 `json:"embeddings"`
}

type EmbeddingResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Value   EmbeddingResponseValue `json:"value"`
}

const baseUrl = "http://localhost:8000"

func Embedding(texts []string) (EmbeddingResponseValue, error) {
	url := strings.Join([]string{baseUrl, "/api/v1/embd/batch"}, "")
	requestBody, err := json.Marshal(EmbeddingRequest{Texts: texts})
	if err != nil {
		return EmbeddingResponseValue{}, err
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		return EmbeddingResponseValue{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return EmbeddingResponseValue{}, fmt.Errorf("http error: status code = %d", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return EmbeddingResponseValue{}, err
	}

	var body EmbeddingResponse
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return EmbeddingResponseValue{}, fmt.Errorf("unable to unmarshal %s: %s", string(bodyBytes), err.Error())
	}

	return body.Value, nil
}
