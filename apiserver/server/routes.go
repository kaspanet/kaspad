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
	routeParamTxID    = "txID"
	routeParamTxHash  = "txHash"
	routeParamAddress = "address"
)

const (
	queryParamSkip  = "skip"
	queryParamLimit = "limit"
)

func makeHandler(handler func(routeParams map[string]string, queryParams map[string][]string, ctx *utils.APIServerContext) (interface{}, *utils.HandlerError)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := utils.ToAPIServerContext(r.Context())
		response, hErr := handler(mux.Vars(r), r.URL.Query(), ctx)
		if hErr != nil {
			sendErr(ctx, w, hErr)
			return
		}
		sendJSONResponse(w, response)
	}
}

func sendErr(ctx *utils.APIServerContext, w http.ResponseWriter, hErr *utils.HandlerError) {
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

func mainHandler(_ map[string]string, _ map[string][]string, _ *utils.APIServerContext) (interface{}, *utils.HandlerError) {
	return "API server is running", nil
}

func addRoutes(router *mux.Router) {
	router.HandleFunc("/", makeHandler(mainHandler))

	router.HandleFunc(
		fmt.Sprintf("/transaction/id/{%s}", routeParamTxID),
		makeHandler(func(routeParams map[string]string, queryParams map[string][]string, ctx *utils.APIServerContext) (interface{}, *utils.HandlerError) {
			return controllers.GetTransactionByIDHandler(routeParams[routeParamTxID])
		})).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transaction/hash/{%s}", routeParamTxHash),
		makeHandler(func(routeParams map[string]string, queryParams map[string][]string, ctx *utils.APIServerContext) (interface{}, *utils.HandlerError) {
			return controllers.GetTransactionByHashHandler(routeParams[routeParamTxHash])
		})).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transactions/address/{%s}", routeParamAddress),
		makeHandler(func(vars map[string]string, ctx *utils.APIServerContext) (interface{}, *utils.HandlerError) {
			return controllers.GetTransactionsByAddressHandler(vars[routeParamAddress])
		})).
		Methods("GET")
}
