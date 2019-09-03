package server

import (
	"encoding/json"
	"fmt"
	"github.com/daglabs/btcd/apiserver/controllers"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	routeParamTxID   = "txID"
	routeParamTxHash = "txHash"
)

func makeHandler(handler func(vars map[string]string, ctx *utils.ApiServerContext) (interface{}, *utils.HandlerError)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := utils.NewAPIServerContext(r.Context())
		response, hErr := handler(mux.Vars(r), ctx)
		if hErr != nil {
			sendErr(ctx, w, hErr)
			return
		}
		sendJSONResponse(w, response)
	}
}

func sendErr(ctx *utils.ApiServerContext, w http.ResponseWriter, hErr *utils.HandlerError) {
	errMsg := fmt.Sprintf("got error: %s", hErr)
	ctx.Warnf(errMsg)
	w.WriteHeader(hErr.ErrorCode)
	sendJSONResponse(w, hErr)
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

func mainHandler(_ map[string]string, _ *utils.ApiServerContext) (interface{}, *utils.HandlerError) {
	return "API server is running", nil
}

func addRoutes(router *mux.Router) {
	router.HandleFunc("/", makeHandler(mainHandler))

	router.HandleFunc(
		fmt.Sprintf("/transaction/id/{%s}", routeParamTxID),
		makeHandler(func(vars map[string]string, ctx *utils.ApiServerContext) (interface{}, *utils.HandlerError) {
			return controllers.GetTransactionByIDHandler(vars[routeParamTxID])
		})).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transaction/hash/{%s}", routeParamTxHash),
		makeHandler(func(vars map[string]string, ctx *utils.ApiServerContext) (interface{}, *utils.HandlerError) {
			return controllers.GetTransactionByHashHandler(vars[routeParamTxHash])
		})).
		Methods("GET")
}
