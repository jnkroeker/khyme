package task

import (
	"net/url"
	"time"

	"github.com/jnkroeker/khyme/business/data/store/task"
)

func CreateTask(url url.URL) task.NewTask {
	return task.NewTask{
		Version:        "v1",
		InputResource:  url,
		OutputResource: url,
		Hooks:          "hooks",
		ExecutionImage: "jnkroeker/processor:0.0.0",
		Timeout:        time.Second * 60,
	}
}
