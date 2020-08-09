package worker

import (
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
)

var directMessageHandlers = map[string]CommandHandler {
	"covid19": covid19Handler,
	"nsfw": nsfwHandler,
}

func getDirectMessageHelp() string {
	mesg := "Available topic:\n"
	for k := range mentionHandlers {
		mesg = mesg + "  " + k + "\n"
	}
	return mesg
}

func (w *Worker) processDirectIM(event *model.MessageEvent) {
	command, subCommand, args := parseManageCommand(event.Resource.Content)
	log.Println("Direct command:", command, subCommand, args)
	if handler, exists := directMessageHandlers[command]; exists {
		log.Printf("Found handler for command [%s]", command)
		go handler(w, command, subCommand, args, event)
	} else {
		utils.ExecuteWithRetry(func() error {
			return w.SendTextMessage(event.GetThreadId(), getDirectMessageHelp())
		})
	}
}