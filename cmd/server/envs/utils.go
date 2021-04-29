package envs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

func toApiObject(env *database.Env) *api.Env {
	// Secret attribute is intentionally omitted
	return &api.Env{Id: env.Id, Name: env.Name, ServerUri: env.ServerUri, Namespace: env.Namespace}
}

func sendResult(entity *database.Env, err error, rw http.ResponseWriter) {
	if err != nil {
		writeErrorResponse(rw, err)
		return
	}

	writeJsonResponse(rw, entity)
}

func sendResults(entities []database.Env, err error, rw http.ResponseWriter) {
	if err != nil {
		writeErrorResponse(rw, err)
		return
	}

	envs := make([]api.Env, len(entities))
	for i, v := range entities {
		envs[i] = api.Env{Id: v.Id, Name: v.Name}
	}

	writeJsonResponse(rw, envs)
}

func writeJsonResponse(resp http.ResponseWriter, v interface{}) {
	enc := json.NewEncoder(resp)
	if err := enc.Encode(v); err != nil {
		log.Printf("error encoding response: %v", err)
		if _, err := resp.Write([]byte("{}")); err != nil {
			log.Printf("error sending empty reponse: %v", err)
		}
	}
}
