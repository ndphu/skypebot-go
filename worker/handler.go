package worker

import (
	"github.com/ndphu/skypebot-go/model"
)

type CommandHandler func(w *Worker, command string, subCommand string, args []string, evt *model.MessageEvent) error
