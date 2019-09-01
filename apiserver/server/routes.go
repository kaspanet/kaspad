package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func makeHandler(handler func(vars map[string]string, ctx *apiServerContext) (interface{}, *handlerError)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := newAPIServerContext(r.Context())
		response, hErr := handler(mux.Vars(r), ctx)
		if hErr != nil {
			sendErr(ctx, w, hErr)
			return
		}
		sendJSONResponse(w, response)
	}
}

func sendJSONResponse(w http.ResponseWriter, response interface{}) {
	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(w, string(b))
	if err != nil {
		panic(err)
	}
}

func mainHandler(vars map[string]string, ctx *apiServerContext) (interface{}, *handlerError) {
	return "API server is running", nil
}

func addRoutes(router *mux.Router) {
	router.HandleFunc("/", makeHandler(mainHandler))
	router.HandleFunc("/transaction/id/{txID}", makeHandler(getTransactionByIDHandler)).Methods("GET")
}
