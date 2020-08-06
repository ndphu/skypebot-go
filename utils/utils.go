package utils

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

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
//
//func ExecuteHttpRequest(req *http.Request) ([]byte, error) {
//	log.Println("Making request to", req.URL.String())
//	client := http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		log.Println("Fail to execute HTTP request", err)
//		return nil, err
//	}
//	defer resp.Body.Close()
//	log.Println(resp.Status)
//	if resp.StatusCode == 429 {
//		return nil, ErrorLimitRequestExceeded
//	}
//	respBody, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		log.Println("Fail to read body", err)
//		return nil, err
//	}
//	log.Println(string(respBody))
//	return respBody, nil
//}
func ExecuteHttpRequestExtended(req *http.Request) (status int, headers http.Header, body []byte, err error) {
	log.Println("Making", req.Method, "request to", req.URL.String())
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Fail to execute HTTP request", err)
		return -1, nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fail to read body", err)
		return resp.StatusCode, resp.Header, nil, err
	}
	return resp.StatusCode, resp.Header, respBody, nil
}

func LogHeaders(headers http.Header)  {
	log.Println("================Headers================")
	for k := range headers {
		log.Println(k, ":", headers.Get(k))
	}
	log.Println("=======================================")
}