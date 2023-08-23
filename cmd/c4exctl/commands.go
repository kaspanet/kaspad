package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/c4ei/yunseokyeol/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.C4exdMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.C4exdMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.C4exdMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.C4exdMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.C4exdMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.C4exdMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.C4exdMessage_BanRequest{}),
	reflect.TypeOf(protowire.C4exdMessage_UnbanRequest{}),
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
