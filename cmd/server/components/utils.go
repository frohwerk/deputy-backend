package components

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJsonResponse(resp http.ResponseWriter, v interface{}) {
	enc := json.NewEncoder(resp)
	if err := enc.Encode(v); err != nil {
		log.Printf("error encoding response: %v", err)
		if _, err := resp.Write([]byte("{}")); err != nil {
			log.Printf("error sending empty reponse: %v", err)
		}
	}
}
