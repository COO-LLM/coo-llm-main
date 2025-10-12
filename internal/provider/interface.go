package provider

import (
	"context"
	"encoding/json"
)

type Provider interface {
	Name() string
	Generate(ctx context.Context, req *Request) (*Response, error)
	ListModels(ctx context.Context) ([]string, error)
}

type Request struct {
	Model  string                 `json:"model"`
	Input  map[string]interface{} `json:"input"`
	APIKey string                 `json:"api_key"`
}

type Response struct {
	RawResponse []byte
	HTTPCode    int
	Err         error
	TokensUsed  int
	Latency     int64 // in milliseconds
}

func (r *Response) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.RawResponse)
}
