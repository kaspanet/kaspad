package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fabbez/topiad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.topiadMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.topiadMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.topiadMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.topiadMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.topiadMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.topiadMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.topiadMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.topiadMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.topiadMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.topiadMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.topiadMessage_BanRequest{}),
	reflect.TypeOf(protowire.topiadMessage_UnbanRequest{}),
}

type commandDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func commandDescriptions() []*commandDescription {
	commandDescriptions := make([]*commandDescription, len(commandTypes))

	for i, commandTypeWrapped := range commandTypes {
		commandType := unwrapCommandType(commandTypeWrapped)

		name := strings.TrimSuffix(commandType.Name(), "RequestMessage")
		numFields := commandType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := commandType.Field(i)

			if !isFieldExported(field) {
				continue
			}

			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		commandDescriptions[i] = &commandDescription{
			name:       name,
			parameters: parameters,
			typeof:     commandTypeWrapped,
		}
	}

	return commandDescriptions
}

func (cd *commandDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(cd.name)
	for _, parameter := range cd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}
