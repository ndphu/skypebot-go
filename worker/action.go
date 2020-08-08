package worker

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/ndphu/skypebot-go/media"
	"github.com/ndphu/skypebot-go/model"
	"image"
	_ "image/jpeg"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (w *Worker) ProcessMessage(evt *model.MessageEvent) error {
	threadId, from := w.parseInfo(evt)
	log.Println("Processing message from", from, "on thread", threadId)
	if evt.Type == "EventMessage" && evt.ResourceType == "NewMessage" && evt.Resource.MessageType == "RichText" {
		if w.isDirectMention(evt) {
			go w.processMention(evt)
		} else if w.isDirectIM(evt) {
			go w.processDirectIM(evt)
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
		w.processCommand(w.normalizeMessageContent(command), event)
	}
}

func (w *Worker) processCommand(commandString string, event *model.MessageEvent) {
	threadId, _ := w.parseInfo(event)
	if commandString == "keyword" || commandString == "keyworks" {
		message := "Supported keywords:\n"
		keywords := media.GetKeywords()
		for i, k := range keywords {
			message = message + k + "    "
			if i != 0 && i%5 == 0 {
				message = message + "\n"
			}
		}
		go w.postWithRetry(func() error {
			return w.PostTextMessage(threadId, message)
		}, 5, 2*time.Second)
		return
	}

	if commandString == "help" || commandString == "usage" {
		usageMessage := "Usage:\n    help: show this message\n    keyword: list supported keywords\n    &lt;keyword&gt; &lt; number of pic - default 1, max 10&gt;"
		go w.postWithRetry(func() error {
			return w.PostTextMessage(threadId, usageMessage)
		}, 5, 2*time.Second)

		return
	}
	r := regexp.MustCompile("^(.*?) ([0-9]*?)$")
	chunks := r.FindAllStringSubmatch(commandString, -1)
	if len(chunks) > 0 {
		keyword := chunks[0][1]
		if count, err := strconv.Atoi(chunks[0][2]); err == nil {
			if count > 10 {
				count = 10
			}
			for i := 0; i < count; i ++ {
				go w.sendRandomImage(threadId, keyword)
			}
		}
	} else {
		go w.sendRandomImage(threadId, commandString)
	}
}

func (w*Worker) parseInfo(evt *model.MessageEvent) (string, string) {
	threadId := path.Base(evt.Resource.ConversationLink)
	from := path.Base(evt.Resource.From)
	return threadId, from
}
//
//func (w *Worker) TakeAction(evt *model.MessageEvent, action Action) {
//	if action.Type == ActionTypeReact {
//		threadId, _ := parseInfo(evt)
//		for _, emotion := range action.Data["emotions"].([]interface{}) {
//			if match, err := regexp.MatchString("^<ss type=\"(.*)\">.*<\\/ss>$", evt.Resource.Content); err != nil {
//				//panic(err)
//			} else {
//				if match {
//					log.Println("Ignore single icon message")
//					continue
//				}
//			}
//			log.Println("Reacting with emotion", emotion)
//			w.ReactMessage(threadId, evt.Resource.Id, emotion.(string))
//		}
//	}
//}
func (w *Worker) sendRandomImage(threadId, keyword string) error {
	if keyword == "" {
		keyword = media.GetKeywords()[0]
	}
	log.Println("Send random image for keyword", keyword, "to thread", threadId)
	var mediaUrl string
	var mediaPayload []byte
	for {
		urls := media.RandomMediaUrl(keyword, 1)
		if len(urls) == 0 {
			break
		}
		mediaUrl = urls[0]
		payload, err := media.DownloadMediaUrl(mediaUrl)
		if err != nil {
			continue
		}
		if len(payload) == 503 {
			continue
		}
		mediaPayload = payload
		break
	}
	if len(mediaPayload) == 0 {
		return nil
	}
	filename := path.Base(mediaUrl)
	transId := uuid.New().String()
	objectId, err := w.CreateObject(threadId, filename, transId)

	if err != nil {
		log.Println("Fail to create object", transId, err)
		return err
	}
	if err := w.UploadObject(objectId, transId, mediaPayload); err != nil {
		log.Println("Fail to upload object", objectId, transId)
		return err
	}
	c, _, e := image.DecodeConfig(bytes.NewReader(mediaPayload))
	var width = 0
	var height = 0
	if e != nil {
		log.Println("fail to decode image config", e)
	} else {
		width = c.Width
		height = c.Height
	}

	return w.postWithRetry(func() error {
		return w.PostImageToThread(threadId,
			objectId,
			filename,
			len(mediaPayload),
			width, height)
	}, 5, 2*time.Second)

}
