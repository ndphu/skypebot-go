package worker

import (
	"github.com/ndphu/skypebot-go/model"
	"path"
)

func (w *Worker) IsDirectIM(event *model.MessageEvent) bool {
	return path.Base(event.Resource.ConversationLink) == path.Base(event.Resource.From)
}

func (w *Worker) IsMessageFromManager(event *model.MessageEvent) bool {
	return w.IsDirectIM(event) && contains(w.managers, path.Base(event.Resource.From))
}

