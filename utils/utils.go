package utils

import (
	"errors"
	"math/rand"
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
