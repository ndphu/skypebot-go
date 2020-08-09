package worker

import (
	"github.com/ndphu/skypebot-go/media"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"path"
	"strconv"
	"strings"
)

var handlers map[string]CommandHandler = map[string]CommandHandler{
	"reload": func(w *Worker, command string, subCommand string, args []string, evt *model.MessageEvent) error {
		if subCommand == "medias" {
			return media.ReloadMedias()
		}
		return nil
	},
	"nsfw":    nsfwHandler,
	"covid19": covid19Handler,
	"conversations": func(w *Worker, command string, subCommand string, args []string, evt *model.MessageEvent) error {
		switch subCommand {
		case "list":
			limit := 10
			if len(args) > 0 {
				if v, err := strconv.Atoi(args[0]); err == nil {
					limit = v
				}
			}
			lines := make([]string, 0)
			if conversations, err := w.GetConversations(limit); err != nil {
				log.Println("Fail to get conversations")
				return utils.ExecuteWithRetry(func() error {
					return w.SendTextMessage(evt.GetThreadId(),
						"Fail to get conversations by error: "+err.Error())
				})
			} else {
				for _, con := range conversations {
					lines = append(lines, strings.Join([]string{con.Topic, con.Id}, "\n"))
				}
				return utils.ExecuteWithRetry(func() error {
					return w.SendTextMessage(evt.GetThreadId(),
						strings.Join(lines, "\n\n"))
				})
			}
			break
		case "enableNsfw":
			count := 0
			for _, threadId := range args {
				w.nsfwEnabledThreads = append(w.nsfwEnabledThreads, threadId)
				count ++
			}
			return utils.ExecuteWithRetry(func() error {
				return w.SendTextMessage(evt.GetThreadId(),
					"completed")
			})
		}
		return nil
	},
}

func (w *Worker) processManageIM(evt *model.MessageEvent) error {
	log.Printf("Processing manage IM: [%s]\n", evt.Resource.Content)
	command, subCommand, args := parseManageCommand(evt.Resource.Content)
	if handler, exists := handlers[command]; exists {
		handler(w, command, subCommand, args, evt)
	} else {
		return w.printManageHelp(evt)
	}

	return nil
}

func (w *Worker) printManageHelp(evt *model.MessageEvent) error {
	helpMessage := "Available commands:\n"
	for k := range handlers {
		helpMessage = helpMessage + "  - " + k + "\n"
	}
	return utils.ExecuteWithRetry(func() error {
		return w.SendTextMessage(path.Base(evt.Resource.ConversationLink), helpMessage)
	})
}

func parseManageCommand(input string) (command string, subCommand string, args []string) {
	normalized := normalizeMessageContent(input)
	normalized = standardizeSpaces(normalized)
	chunks := strings.Split(normalized, " ")
	if len(chunks) > 0 {
		command = chunks[0]
	}
	if len(chunks) > 1 {
		subCommand = chunks[1]
	}
	if len(chunks) > 2 {
		for i := 2; i < len(chunks); i ++ {
			args = append(args, chunks[i])
		}
	}
	return command, subCommand, args
}
