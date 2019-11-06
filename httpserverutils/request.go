package httpserverutils

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

// HandlerFunc is a handler function that is passed to the
// MakeHandler wrapper and gets the relevant request fields
// from it.
type HandlerFunc func(ctx *ServerContext, r *http.Request, routeParams map[string]string, queryParams map[string]string, requestBody []byte) (
	interface{}, error)

// MakeHandler is a wrapper function that takes a handler in the form of HandlerFunc
// and returns a function that can be used as a handler in mux.Router.HandleFunc.
func MakeHandler(handler HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := ToServerContext(r.Context())

		var requestBody []byte
		if r.Method == "POST" {
			var err error
			requestBody, err = ioutil.ReadAll(r.Body)
			if err != nil {
				SendErr(ctx, w, errors.New("Error reading POST data"))
			}
		}

		flattenedQueryParams, err := flattenQueryParams(r.URL.Query())
		if err != nil {
			SendErr(ctx, w, err)
			return
		}

		response, err := handler(ctx, r, mux.Vars(r), flattenedQueryParams, requestBody)
		if err != nil {
			SendErr(ctx, w, err)
			return
		}
		if response != nil {
			SendJSONResponse(w, response)
		}
	}
}

func flattenQueryParams(queryParams map[string][]string) (map[string]string, error) {
	flattenedMap := make(map[string]string)
	for param, valuesSlice := range queryParams {
		if len(valuesSlice) > 1 {
			return nil, NewHandlerError(http.StatusUnprocessableEntity, errors.Errorf("Couldn't parse the '%s' query parameter:"+
				" expected a single value but got multiple values", param))
		}
		flattenedMap[param] = valuesSlice[0]
	}
	return flattenedMap, nil
}
