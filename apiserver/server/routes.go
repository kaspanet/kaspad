package server

import (
	"fmt"
	"github.com/daglabs/btcd/httpserverutils"
	"net/http"
	"strconv"

	"github.com/daglabs/btcd/apiserver/controllers"
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
	queryParamOrder = "order"
)

const (
	defaultGetTransactionsLimit = 100
	defaultGetBlocksLimit       = 25
	defaultGetBlocksOrder       = controllers.OrderAscending
)

func mainHandler(_ *httpserverutils.ServerContext, _ *http.Request, _ map[string]string, _ map[string]string, _ []byte) (interface{}, *httpserverutils.HandlerError) {
	return struct {
		Message string `json:"message"`
	}{
		Message: "API server is running",
	}, nil
}

func addRoutes(router *mux.Router) {
	router.HandleFunc("/", httpserverutils.MakeHandler(mainHandler))

	router.HandleFunc(
		fmt.Sprintf("/transaction/id/{%s}", routeParamTxID),
		httpserverutils.MakeHandler(getTransactionByIDHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transaction/hash/{%s}", routeParamTxHash),
		httpserverutils.MakeHandler(getTransactionByHashHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/transactions/address/{%s}", routeParamAddress),
		httpserverutils.MakeHandler(getTransactionsByAddressHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/utxos/address/{%s}", routeParamAddress),
		httpserverutils.MakeHandler(getUTXOsByAddressHandler)).
		Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("/block/{%s}", routeParamBlockHash),
		httpserverutils.MakeHandler(getBlockByHashHandler)).
		Methods("GET")

	router.HandleFunc(
		"/blocks",
		httpserverutils.MakeHandler(getBlocksHandler)).
		Methods("GET")

	router.HandleFunc(
		"/fee-estimates",
		httpserverutils.MakeHandler(getFeeEstimatesHandler)).
		Methods("GET")

	router.HandleFunc(
		"/transaction",
		httpserverutils.MakeHandler(postTransactionHandler)).
		Methods("POST")
}

func convertQueryParamToInt(queryParams map[string]string, param string, defaultValue int) (int, *httpserverutils.HandlerError) {
	if _, ok := queryParams[param]; ok {
		intValue, err := strconv.Atoi(queryParams[param])
		if err != nil {
			return 0, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter: %s", param, err))
		}
		return intValue, nil
	}
	return defaultValue, nil
}

func getTransactionByIDHandler(_ *httpserverutils.ServerContext, _ *http.Request, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	return controllers.GetTransactionByIDHandler(routeParams[routeParamTxID])
}

func getTransactionByHashHandler(_ *httpserverutils.ServerContext, _ *http.Request, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	return controllers.GetTransactionByHashHandler(routeParams[routeParamTxHash])
}

func getTransactionsByAddressHandler(_ *httpserverutils.ServerContext, _ *http.Request, routeParams map[string]string, queryParams map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	skip, hErr := convertQueryParamToInt(queryParams, queryParamSkip, 0)
	if hErr != nil {
		return nil, hErr
	}
	limit, hErr := convertQueryParamToInt(queryParams, queryParamLimit, defaultGetTransactionsLimit)
	if hErr != nil {
		return nil, hErr
	}
	if _, ok := queryParams[queryParamLimit]; ok {
		var err error
		skip, err = strconv.Atoi(queryParams[queryParamLimit])
		if err != nil {
			return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
				fmt.Sprintf("Couldn't parse the '%s' query parameter: %s", queryParamLimit, err))
		}
	}
	return controllers.GetTransactionsByAddressHandler(routeParams[routeParamAddress], uint64(skip), uint64(limit))
}

func getUTXOsByAddressHandler(_ *httpserverutils.ServerContext, _ *http.Request, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	return controllers.GetUTXOsByAddressHandler(routeParams[routeParamAddress])
}

func getBlockByHashHandler(_ *httpserverutils.ServerContext, _ *http.Request, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	return controllers.GetBlockByHashHandler(routeParams[routeParamBlockHash])
}

func getFeeEstimatesHandler(_ *httpserverutils.ServerContext, _ *http.Request, _ map[string]string, _ map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	return controllers.GetFeeEstimatesHandler()
}

func getBlocksHandler(_ *httpserverutils.ServerContext, _ *http.Request, _ map[string]string, queryParams map[string]string,
	_ []byte) (interface{}, *httpserverutils.HandlerError) {

	skip, hErr := convertQueryParamToInt(queryParams, queryParamSkip, 0)
	if hErr != nil {
		return nil, hErr
	}
	limit, hErr := convertQueryParamToInt(queryParams, queryParamLimit, defaultGetBlocksLimit)
	if hErr != nil {
		return nil, hErr
	}
	order := defaultGetBlocksOrder
	if orderParamValue, ok := queryParams[queryParamOrder]; ok {
		if orderParamValue != controllers.OrderAscending && orderParamValue != controllers.OrderDescending {
			return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("'%s' is not a valid value for the '%s' query parameter", orderParamValue, queryParamLimit))
		}
		order = orderParamValue
	}
	return controllers.GetBlocksHandler(order, uint64(skip), uint64(limit))
}

func postTransactionHandler(_ *httpserverutils.ServerContext, _ *http.Request, _ map[string]string, _ map[string]string,
	requestBody []byte) (interface{}, *httpserverutils.HandlerError) {
	return nil, controllers.PostTransaction(requestBody)
}
