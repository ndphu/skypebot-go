package utils

import (
	"errors"
	"log"
	"math/rand"
	"strings"
	"time"
)

//
//import (
//	"errors"
//	"io/ioutil"
//	"log"
//	"math/rand"
//	"net/http"
//	"time"
//)
//
var letterRunes = []rune("0123456789")

var ErrorLimitRequestExceeded = errors.New("login rate limit exceeded")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetUTCNow() string {
	utc, _ := time.LoadLocation("UTC")
	layout := "2006-01-02T15:04:05.000Z"
	return time.Now().In(utc).Format(layout)
}

func CompleteThreadId(id string) string {
	threadId := id
	if !strings.HasPrefix(threadId, "19:") {
		threadId = "19:" + threadId
	}
	if !strings.HasSuffix(threadId, "@thread.skype") {
		threadId = threadId + "@thread.skype"
	}
	return threadId
}

func CompleteUserId(id string) (string) {
	userId := id
	if !strings.HasPrefix(userId, "8:") {
		userId = "8:" + userId
	}
	return userId
}

func NormalizeMessageContent(content string) string {
	commandString := strings.TrimPrefix(content, "-")
	commandString = strings.TrimSpace(commandString)
	commandString = strings.ToLower(commandString)
	log.Println("Normalized command:", commandString)
	return commandString
}

