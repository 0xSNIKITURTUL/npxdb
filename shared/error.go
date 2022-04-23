package shared

import (
	"log"
	"net/http"
)

func HandleServerError(writer *http.ResponseWriter, err error) {
	defer log.Println(err.Error())
	(*writer).Write([]byte("Internal Server Error"))
}
