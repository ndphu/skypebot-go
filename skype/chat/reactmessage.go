package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ndphu/skypebot-go/config"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"net/http"
	"time"
)

type emotionRequest struct {
	Emotions string `json:"emotions"`
}

func ReactMessage(conversationId string, messageId string, emotion string) error {
	log.Println("Reacting message", messageId, "in thread", conversationId, "with emotion", emotion, "...")
	urlPath := fmt.Sprintf("/v1/users/ME/conversations/%s/messages/%s/properties?name=emotions", conversationId, messageId)
	emotions, _ := json.Marshal(map[string]string{"key": emotion})
	er := emotionRequest{
		Emotions: string(emotions),
	}
	payload, _ := json.Marshal(er)

	req, _ := http.NewRequest("PUT", config.Get().MessageBaseUrl()+urlPath, bytes.NewReader(payload))
	utils.SetRequestHeaders(req)
	_, _, _, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		return err
	}
	log.Println("React successfully")
	return nil
}
func ReactThread(target, emotion string) error {
	messages, err := GetAllTextMessagesWithLimitAndTimeout(target, 1000)
	log.Println(len(messages))
	for _, m := range messages {
		log.Println(m.Content)
	}
	if err != nil {
		log.Println("Fail to load message in thread", target)
		return err
	}

	for _, msg := range messages {
		if msg.Type != "Message" {
			continue
		}
		for {
			log.Println("Reacting message:", msg.Content)
			if err := ReactMessage(target, msg.Id, emotion); err != nil {
				if err == utils.ErrorLimitRequestExceeded {
					log.Println("Got 429 error. Sleeping for 30 second and retry...")
					time.Sleep(30 * time.Second)
				} else {
					log.Println("Fail to react message", msg.Id, "in thread", target, "with emotion", emotion, err)
					return err
				}
			} else {
				break
			}
		}
		time.Sleep(15 * time.Second)
	}
	return nil
}
