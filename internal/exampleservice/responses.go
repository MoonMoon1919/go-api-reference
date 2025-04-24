package exampleservice

import "github.com/moonmoon1919/go-api-reference/pkg/example"

type PatchExampleResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

func NewPatchExampleResponseFromExample(e example.Example) PatchExampleResponse {
	return PatchExampleResponse{
		Id:      e.Id,
		Message: e.Message,
	}
}

type CreateExampleResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

func NewCreateExampleResponseFromExample(e example.Example) CreateExampleResponse {
	return CreateExampleResponse{
		Id:      e.Id,
		Message: e.Message,
	}
}

type GetExampleResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

func NewGetExampleResponseFromExample(e example.Example) GetExampleResponse {
	return GetExampleResponse{
		Id:      e.Id,
		Message: e.Message,
	}
}

type ListExampleResponse struct {
	Items []GetExampleResponse `json:"items"`
}

func NewListExampleResponseFromExamples(e []example.Example) ListExampleResponse {
	items := make([]GetExampleResponse, len(e))

	for idx, i := range e {
		items[idx] = NewGetExampleResponseFromExample(i)
	}

	return ListExampleResponse{
		Items: items,
	}
}
