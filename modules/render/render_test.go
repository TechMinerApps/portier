package render

import (
	"testing"

	"github.com/TechMinerApps/portier/models"
	"github.com/mmcdole/gofeed"
)

func Test_renderer_Render(t *testing.T) {
	type fields struct {
		Config Config
	}
	type args struct {
		feed *models.Feed
	}

	Template := "{{ .Item.Title }}"
	feed := &models.Feed{
		SourceID: 0,
		FeedID:   "",
		Item: &gofeed.Item{
			Title:       "Unit Test is Great!",
			Description: "",
			Content:     "",
			Link:        "",
			Updated:     "",
			Author:      &gofeed.Person{},
			GUID:        "",
			Image:       &gofeed.Image{},
		},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Normal",
			fields: fields{
				Config: Config{
					Template: Template,
				},
			},
			args: args{
				feed: feed,
			},
			want:    "Unit Test is Great!",
			wantErr: false,
		},
		{
			name: "Error",
			fields: fields{
				Config: Config{
					Template: "{{.NotExists}}",
				},
			},
			args: args{
				feed: feed,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := NewRenderer(tt.fields.Config)
			got, err := r.Render(tt.args.feed)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderer.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("renderer.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}
