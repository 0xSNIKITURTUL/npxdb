package shared

import (
	"log"
	"net/http"
)

func HandleServerError(writer *http.ResponseWriter, err error) {
	defer log.Fatalln(err.Error())
	(*writer).Write([]byte("Internal Server Error"))
}
