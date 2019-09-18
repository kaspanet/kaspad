package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	queryParamOrder = "order"
)

const (
	defaultGetTransactionsLimit = 100
	defaultGetBlocksLimit       = 25
	defaultGetBlocksOrder       = controllers.OrderAscending
)

type handlerFunc func(ctx *utils.APIServerContext, routeParams map[string]string, queryParams map[string]string, requestBody []byte) (
	interface{}, *utils.HandlerError)

func makeHandler(handler handlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := utils.ToAPIServerContext(r.Context())

		var requestBody []byte
		if r.Method == "POST" {
			var err error
			requestBody, err = ioutil.ReadAll(r.Body)
			if err != nil {
				sendErr(ctx, w, utils.NewHandlerError(500, "Internal server error occured"))
			}
		}

		flattenedQueryParams, hErr := flattenQueryParams(r.URL.Query())
		if hErr != nil {
			sendErr(ctx, w, hErr)
			return
		}

		response, hErr := handler(ctx, mux.Vars(r), flattenedQueryParams, requestBody)
		if hErr != nil {
			sendErr(ctx, w, hErr)
			return
		}
		if response != nil {
			sendJSONResponse(w, response)
		}
	}
}

func flattenQueryParams(queryParams map[string][]string) (map[string]string, *utils.HandlerError) {
	flattenedMap := make(map[string]string)
	for param, valuesSlice := range queryParams {
		if len(valuesSlice) > 1 {
			return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter:"+
				" expected a single value but got multiple values", param))
		}
		flattenedMap[param] = valuesSlice[0]
	}
	return flattenedMap, nil
}

type clientError struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func sendErr(ctx *utils.APIServerContext, w http.ResponseWriter, hErr *utils.HandlerError) {
	errMsg := fmt.Sprintf("got error: %s", hErr)
	ctx.Warnf(errMsg)
	w.WriteHeader(hErr.Code)
	sendJSONResponse(w, &clientError{
		ErrorCode:    hErr.Code,
		ErrorMessage: hErr.ClientMessage,
	})
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

func mainHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string]string, _ []byte) (interface{}, *utils.HandlerError) {
	return struct {
		Message string `json:"message"`
	}{
		Message: "API server is running",
	}, nil
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
		"/blocks",
		makeHandler(getBlocksHandler)).
		Methods("GET")

	router.HandleFunc(
		"/fee-estimates",
		makeHandler(getFeeEstimatesHandler)).
		Methods("GET")

	router.HandleFunc(
		"/transaction",
		makeHandler(postTransactionHandler)).
		Methods("POST")
}

func convertQueryParamToInt(queryParams map[string]string, param string, defaultValue int) (int, *utils.HandlerError) {
	if _, ok := queryParams[param]; ok {
		intValue, err := strconv.Atoi(queryParams[param])
		if err != nil {
			return 0, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Couldn't parse the '%s' query parameter: %s", param, err))
		}
		return intValue, nil
	}
	return defaultValue, nil
}

func getTransactionByIDHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

	return controllers.GetTransactionByIDHandler(routeParams[routeParamTxID])
}

func getTransactionByHashHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

	return controllers.GetTransactionByHashHandler(routeParams[routeParamTxHash])
}

func getTransactionsByAddressHandler(_ *utils.APIServerContext, routeParams map[string]string, queryParams map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

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
			return nil, utils.NewHandlerError(http.StatusUnprocessableEntity,
				fmt.Sprintf("Couldn't parse the '%s' query parameter: %s", queryParamLimit, err))
		}
	}
	return controllers.GetTransactionsByAddressHandler(routeParams[routeParamAddress], uint64(skip), uint64(limit))
}

func getUTXOsByAddressHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

	return controllers.GetUTXOsByAddressHandler(routeParams[routeParamAddress])
}

func getBlockByHashHandler(_ *utils.APIServerContext, routeParams map[string]string, _ map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

	return controllers.GetBlockByHashHandler(routeParams[routeParamBlockHash])
}

func getFeeEstimatesHandler(_ *utils.APIServerContext, _ map[string]string, _ map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

	return controllers.GetFeeEstimatesHandler()
}

func getBlocksHandler(_ *utils.APIServerContext, _ map[string]string, queryParams map[string]string,
	_ []byte) (interface{}, *utils.HandlerError) {

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
			return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("'%s' is not a valid value for the '%s' query parameter", orderParamValue, queryParamLimit))
		}
		order = orderParamValue
	}
	return controllers.GetBlocksHandler(order, uint64(skip), uint64(limit))
}

func postTransactionHandler(_ *utils.APIServerContext, _ map[string]string, _ map[string]string,
	requestBody []byte) (interface{}, *utils.HandlerError) {
	return nil, controllers.PostTransaction(requestBody)
}
