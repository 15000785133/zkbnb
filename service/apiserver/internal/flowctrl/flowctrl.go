package flowctrl

import "net/http"

func FlowControlHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		// If not forbidden
		// continue to do the next process
		next(writer, request)
	}
}
