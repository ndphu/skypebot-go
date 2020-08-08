package worker

import (
	"github.com/ndphu/skypebot-go/model"
	"path"
	"strings"
)

func (w *Worker) isDirectIM(event *model.MessageEvent) bool {
	return path.Base(event.Resource.ConversationLink) == path.Base(event.Resource.From)
}
func (w *Worker) processDirectIM(event *model.MessageEvent) {
	content := event.Resource.Content
	commandString := w.normalizeMessageContent(content)
	w.processCommand(commandString, event)
}

func (w *Worker) normalizeMessageContent(content string) string {
	commandString := strings.ReplaceAll(content, "-", "")
	commandString = strings.TrimSpace(commandString)
	commandString = strings.ToLower(commandString)
	return commandString
}

