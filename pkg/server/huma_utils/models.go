package huma_utils

type JsonBody[T any] struct {
	Body T
}

func NewJsonBody[T any](v T) *JsonBody[T] {
	return &JsonBody[T]{Body: v}
}

type IdByPath struct {
	Id int64 `path:"id"`
}

type StringIdByPath struct {
	Id string `path:"id"`
}

type Empty struct {
	Body map[string]any
}

type List[T any] struct {
	Body ListBody[T]
}

type ListBody[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"total_count"`
}

func NewList[T any](l []T, totalCount int) *List[T] {
	if l == nil {
		l = []T{}
	}
	return &List[T]{
		Body: ListBody[T]{
			Items:      l,
			TotalCount: totalCount,
		},
	}
}
