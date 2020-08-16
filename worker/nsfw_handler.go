package worker

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/ndphu/skypebot-go/media"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"image"
	"log"
	"path"
	"strconv"
	"strings"
)

var nsfwHandler CommandHandler = func(w *Worker, command string, subCommand string, args [] string, evt *model.MessageEvent, ) error {
	if !contains(w.nsfwEnabledThreads, evt.GetThreadId()) {
		go utils.ExecuteWithRetry(func() error {
			return w.SendTextMessage(evt.GetThreadId(),
				"NSFW content is not enabled for this conversation.\n" +
				"Please contact BOT manager for more information. Thank you.")
		})
		return nil
	}
	if subCommand == "" || subCommand == "keyword" || !contains(media.GetKeywords(), subCommand) {
		go utils.ExecuteWithRetry(func() error {
			return w.SendTextMessage(evt.GetThreadId(),
				"<b>Available keywords:</b>\n"+strings.Join(media.GetKeywords(), ", "))

		})
		return nil
	}
	keyword := subCommand
	count := 1
	if len(args) > 0 {
		if value, err := strconv.Atoi(args[0]); err == nil {
			if value > 10 {
				count = 10
			} else if value < 0 {
				count = 0
			} else {
				count = value
			}
		}
	}
	for i := 0; i < count; i ++ {
		go w.SendRandomImage(evt.GetThreadId(), keyword)
	}
	return nil
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func (w *Worker) SendRandomImage(threadId, keyword string) error {
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

	return utils.ExecuteWithRetry(func() error {
		return w.PostImageToThread(threadId,
			objectId,
			filename,
			len(mediaPayload),
			width, height)
	})

}
