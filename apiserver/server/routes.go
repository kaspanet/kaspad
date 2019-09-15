package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daglabs/btcd/apiserver/controllers"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/gorilla/mux"
)

const (
	routeParamTxID      = "txID"
	routeParamTxHash    = "txHash"
	routeParamAddress   = "address"
	routeParamBlockHash = "blockHash"
)

const (
	queryParamSkip  = "skip"
	queryParamLimit = "limit"
)

const defaultGetTransactionsLimit = 100

func makeHandler(
	handler func(ctx *utils.APIServerContext, routeParams map[string]string, queryParams map[string][]string) (
		interface{}, *utils.HandlerError)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := utils.ToAPIServerContext(r.Context())
		response, hErr := handler(ctx, mux.Vars(r), r.URL.Query())
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

func mainHandler(_ *utils.APIServerContext, _ map[string]string, _ map[string][]string) (interface{}, *utils.HandlerError) {
	return "API server is running", nil
}

func addRoutes(router *mux.Router) {
	router.HandleFunc("/", makeHandler(mainHandler))

	router.HandleFunc(
		fmt.Sprintf("/transaction/id/{%s}", routeParamTxID),
		makeHandler(getTransactionByIDHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transaction/hash/{%s}", routeParamTxHash),
		makeHandler(getTransactionByHashHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transactions/address/{%s}", routeParamAddress),
		makeHandler(getTransactionsByAddressHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/utxos/address/{%s}", routeParamAddress),
		makeHandler(getUTXOsByAddressHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/block/{%s}", routeParamBlockHash),
		makeHandler(getBlockByHashHandler)).
		Methods("GET")

	router.HandleFunc(
		"/fee-estimates",
		makeHandler(getFeeEstimatesHandler)).
		Methods("GET")
}

func getTransactionByIDHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string][]string) (interface{}, *utils.HandlerError) {
	return controllers.GetTransactionByIDHandler(routeParams[routeParamTxID])
}

func getTransactionByHashHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string][]string) (interface{}, *utils.HandlerError) {
	return controllers.GetTransactionByHashHandler(routeParams[routeParamTxHash])
}

func getTransactionsByAddressHandler(_ *utils.APIServerContext, routeParams map[string]string, queryParams map[string][]string) (interface{}, *utils.HandlerError) {
	skip := 0
	limit := defaultGetTransactionsLimit
	if len(queryParams[queryParamSkip]) > 1 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter:"+
			" expected a single value but got an array", queryParamSkip))
	}
	if len(queryParams[queryParamSkip]) == 1 {
		var err error
		skip, err = strconv.Atoi(queryParams[queryParamSkip][0])
		if err != nil {
			return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter: %s", queryParamSkip, err))
		}
	}
	if len(queryParams[queryParamLimit]) > 1 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter:"+
			" expected a single value but got an array", queryParamLimit))
	}
	if len(queryParams[queryParamLimit]) == 1 {
		var err error
		skip, err = strconv.Atoi(queryParams[queryParamLimit][0])
		if err != nil {
			return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter: %s", queryParamLimit, err))
		}
	}
	return controllers.GetTransactionsByAddressHandler(routeParams[routeParamAddress], uint64(skip), uint64(limit))
}

func getUTXOsByAddressHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string][]string) (interface{}, *utils.HandlerError) {
	return controllers.GetUTXOsByAddressHandler(routeParams[routeParamAddress])
}

func getBlockByHashHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string][]string) (interface{}, *utils.HandlerError) {
	return controllers.GetBlockByHashHandler(routeParams[routeParamBlockHash])
}

func getFeeEstimatesHandler(_ *utils.APIServerContext, _ map[string]string, _ map[string][]string) (interface{}, *utils.HandlerError) {
	return controllers.GetFeeEstimatesHandler()
}
