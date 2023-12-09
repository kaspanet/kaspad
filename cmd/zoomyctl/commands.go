package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.ZoomydMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.ZoomydMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.ZoomydMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.ZoomydMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.ZoomydMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.ZoomydMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.ZoomydMessage_BanRequest{}),
	reflect.TypeOf(protowire.ZoomydMessage_UnbanRequest{}),
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
