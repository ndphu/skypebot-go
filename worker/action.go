package worker

import (
	"github.com/ndphu/skypebot-go/media"
	"github.com/ndphu/skypebot-go/model"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func (w *Worker) ProcessMessage(evt *model.MessageEvent) error {
	threadId, from := parseInfo(evt)
	log.Println("Processing message from", from, "on thread", threadId)
	if evt.Type == "EventMessage" && evt.ResourceType == "NewMessage" && evt.Resource.MessageType == "RichText" {
		if w.isDirectMention(evt) {
			w.processMention(evt)
		}
	} else {
	}

	return nil
}

func (w *Worker) getMentionPrefix() (string) {
	return "<at id=\"8:" + w.skypeId + "\">"
}

func (w *Worker) isDirectMention(evt *model.MessageEvent) bool {
	return strings.HasPrefix(strings.TrimSpace(evt.Resource.Content), w.getMentionPrefix())
}
func (w *Worker) processMention(event *model.MessageEvent) {
	r := regexp.MustCompile(`^<at id="(.*?)">(.*?)</at>(.*?)$`)
	founds := r.FindAllStringSubmatch(event.Resource.Content, -1)
	if len(founds) > 0 {
		mention := string(founds[0][1])
		name := string(founds[0][2])
		command := string(founds[0][3])
		log.Println("mention", mention, "name", name, "command", command)
		trimmed := strings.TrimSpace(command)
		w.processCommand(trimmed, event)
	}
}

func (w *Worker) processCommand(trimmed string, event *model.MessageEvent) {
	if trimmed == "keyword" || trimmed == "keyworks" {
		threadId, _ := parseInfo(event)
		w.PostTextMessage(threadId, strings.Join(media.GetCategories(), "\n"))
		return
	}
	if trimmed == "random" {
		w.sendRandomImage("", event)
		return
	}

	r := regexp.MustCompile("^random ([0-9]*?)$")
	chunks := r.FindAllStringSubmatch(trimmed, -1)
	if len(chunks) > 0 {
		log.Println(chunks)
		if count, err := strconv.Atoi(chunks[0][1]); err == nil {
			for i := 0; i < count; i ++ {
				go func() {
					w.sendRandomImage("", event)
				}()
			}
		}
	} else {
		r := regexp.MustCompile("^(.*?) ([0-9]*?)$")
		chunks := r.FindAllStringSubmatch(trimmed, -1)
		if len(chunks) > 0 {
			log.Println(chunks)
			if count, err := strconv.Atoi(chunks[0][2]); err == nil {
				for i := 0; i < count; i ++ {
					go func() {
						w.sendRandomImage(chunks[0][1], event)
					}()
				}
			}
		} else {
			go func() {
				w.sendRandomImage(trimmed, event)
			}()
		}
	}
}

func parseInfo(evt *model.MessageEvent) (string, string) {
	threadId := path.Base(evt.Resource.ConversationLink)
	from := path.Base(evt.Resource.From)
	return threadId, from
}

func (w *Worker) TakeAction(evt *model.MessageEvent, action Action) {
	if action.Type == ActionTypeReact {
		threadId, _ := parseInfo(evt)
		for _, emotion := range action.Data["emotions"].([]interface{}) {
			if match, err := regexp.MatchString("^<ss type=\"(.*)\">.*<\\/ss>$", evt.Resource.Content); err != nil {
				//panic(err)
			} else {
				if match {
					log.Println("Ignore single icon message")
					continue
				}
			}
			log.Println("Reacting with emotion", emotion)
			w.ReactMessage(threadId, evt.Resource.Id, emotion.(string))
		}
	}
}
