package task

import (
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/jnkroeker/khyme/business/data/store/task"
)

type Template struct {
	Name   string
	Create func(resource url.URL) *task.Task
}

type Templater struct {
	templates []Template
	version   string
}

func NewTemplater(templates []Template, version string) Templater {
	return Templater{templates, version}
}

func (t Templater) Create(resource url.URL) *task.Task {
	for _, template := range t.templates {
		if task := template.Create(resource); task != nil {
			task.Version = t.version
			return task
		}
	}
	return nil
}

var Mp4 = &Template{
	Name: "Mp4",
	Create: func(resource url.URL) *task.Task {
		if strings.ToLower(path.Ext(resource.Path)) != ".mp4" {
			return nil
		}

		outUrl := resource
		outUrl.Path = path.Join(os.Getenv("CH_TEMPLATE_MP4_MIRROR_PREFIX"), outUrl.Host, outUrl.Path) + "/"
		outUrl.Host = os.Getenv("CH_TEMPLATE_MP4_MIRROR_BUCKET")

		return &task.Task{
			InputResource:  resource,
			OutputResource: outUrl,
			Hooks:          "mp4",
			ExecutionImage: "jnkroeker/mp4_processor:0.1.4",
			Timeout:        time.Duration(48) * time.Hour,
		}
	},
}
