package static

import (
	"errors"

	"github.com/goccy/go-json"
)

type StaticContentI interface {
	Get() ([]byte, error)
}

type staticContent struct {
	Text string `json:"text"`
}

func NewStaticContent() StaticContentI {
	return &staticContent{
		Text: "Hello world",
	}
}

func (h *staticContent) Get() ([]byte, error) {
	content, err := json.Marshal(h)
	if err != nil {
		err = errors.New("Response marshal error: " + err.Error())
	}
	return content, err
}
