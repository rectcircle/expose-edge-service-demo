package helper

import (
	"fmt"
	"net/http"
)

func RespString(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	fmt.Fprint(w, msg)
}
