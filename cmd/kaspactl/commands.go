package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.KaspadMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetUtxosByAddressesRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_BanRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_UnbanRequest{}),
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
