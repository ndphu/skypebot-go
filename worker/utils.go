package worker

import (
	"log"
	"net/http"
)

func logHeaders(headers http.Header)  {
	log.Println("================Headers================")
	for k := range headers {
		log.Println(k, ":", headers.Get(k))
	}
	log.Println("=======================================")
}