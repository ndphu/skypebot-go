package worker

import (
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"regexp"
	"strings"
)

var mentionHandlers = map[string]CommandHandler{
	"covid19": covid19Handler,
	"nsfw":    nsfwHandler,
}

func (w *Worker) getMentionPrefix() (string) {
	return "<at id=\"8:" + w.skypeId + "\">"
}

func (w *Worker) IsDirectMention(evt *model.MessageEvent) bool {
	return strings.HasPrefix(strings.TrimSpace(evt.Resource.Content), w.getMentionPrefix())
}

func getMentionHelpMessage() string {
	mesg := "Available topic:\n"
	for k := range mentionHandlers {
		mesg = mesg + "  " + k + "\n"
	}
	return mesg
}

func (w *Worker) ProcessMention(event *model.MessageEvent) {
	mentionText := w.splitCommandInMention(event)
	command, subCommand, args := parseManageCommand(mentionText)
	log.Println("Mention command:", command, subCommand, args)
	if handler, exists := mentionHandlers[command]; exists {
		log.Printf("Found handler for command [%s]", command)
		go handler(w, command, subCommand, args, event)
	} else {
		utils.ExecuteWithRetry(func() error {
			return w.SendTextMessage(event.GetThreadId(), getMentionHelpMessage())
		})
	}
}

func (w *Worker) splitCommandInMention(event *model.MessageEvent) string {
	r := regexp.MustCompile(`^<at id="(.*?)">(.*?)</at>(.*?)$`)
	founds := r.FindAllStringSubmatch(event.Resource.Content, -1)
	if len(founds) > 0 {
		mention := string(founds[0][1])
		name := string(founds[0][2])
		command := string(founds[0][3])
		log.Println("mention", mention, "name", name, "command", command)
		return command
	}
	return ""
}
