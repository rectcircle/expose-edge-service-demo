package edgeservice

import (
	"fmt"
	"log"
	"net/http"
)

func Run(serviceID string, port int) {
	log.Printf("[edge service][service %s] start listening on port %d", serviceID, port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[edge service] service id %s, port is %d: url=%s, header=%+v", serviceID, port, r.URL.String(), r.Header)
		w.Write([]byte(fmt.Sprintf("Hello, world! service id is %s,  port is %d", serviceID, port)))
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
