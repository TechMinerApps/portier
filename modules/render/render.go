package render

import (
	"bytes"
	"text/template"

	"github.com/TechMinerApps/portier/models"
)

type Config struct {
	Template string
}

// Renderer is a interface that provide render to text
type Renderer interface {
	Render(feed *models.Feed) (string, error)
}

type renderer struct {
	template *template.Template
}

// NewRenderer return a renderer according to config
func NewRenderer(c Config) (Renderer, error) {
	var r renderer
	var err error
	r.template, err = template.New("render").Parse(c.Template)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *renderer) Render(feed *models.Feed) (string, error) {
	var buffer bytes.Buffer
	if err := r.template.Execute(&buffer, feed.Item); err != nil {
		return "", err
	}
	return buffer.String(), nil

}
