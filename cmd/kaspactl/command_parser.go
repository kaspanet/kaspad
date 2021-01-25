package main

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
)

func parseCommand(args []string, requestDescriptions map[string]*requestDescription) (*protowire.KaspadMessage, error) {
	commandName, parameters := args[0], args[1]

	requestDesc, ok := requestDescriptions[commandName]
	if !ok {
		return nil, errors.Errorf("unknown command: %s", commandName)
	}
	if len(parameters) != len(requestDesc.parameters) {
		return nil, errors.Errorf("command '%s' expects %d parameters but got %d",
			commandName, len(requestDesc.parameters), len(parameters))
	}

	payloadValue := reflect.New(requestDesc.typeof)
	for i, parameterDesc := range requestDesc.parameters {
		field := payloadValue.FieldByName(parameterDesc.name)
		arg := args[i]
		err := setField(field, parameterDesc, arg)
		if err != nil {
			return nil, err
		}
	}

	message := payloadValue.Interface().(appmessage.Message)
	return protowire.FromAppMessage(message)
}

func setField(field reflect.Value, parameterDesc *parameterDescription, valueStr string) error {
	value, err := stringToValue(field, parameterDesc, valueStr)
	if err != nil {
		return err
	}

	field.Set(value)
	return nil
}

func stringToValue(field reflect.Value, parameterDesc *parameterDescription, valueStr string) (reflect.Value, error) {
	var value interface{}
	var err error
	switch field.Kind() {
	case reflect.Bool:
		value, err = strconv.ParseBool(valueStr)
		if err != nil {
			return reflect.Value{}, nil
		}
	case reflect.Int8:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		value = int8(valueInt64)
	case reflect.Int16:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		value = int16(valueInt64)
	case reflect.Int32:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		value = int32(valueInt64)
	case reflect.Int64:
		value, err = strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
	case reflect.Uint8:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		value = uint8(valueUInt64)
	case reflect.Uint16:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		value = uint16(valueUInt64)
	case reflect.Uint32:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		value = uint32(valueUInt64)
	case reflect.Uint64:
		value, err = strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
	case reflect.Float32:
		var valueFloat64 float64
		valueFloat64, err = strconv.ParseFloat(valueStr, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		value = float32(valueFloat64)
	case reflect.Float64:
		value, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return reflect.Value{}, err
		}
	case reflect.String:
		value = valueStr
	case reflect.Struct:
		fieldInterface := field.Interface()
		err := json.Unmarshal([]byte(valueStr), fieldInterface)
		if err != nil {
			return reflect.Value{}, err
		}
		value = fieldInterface
	case reflect.Ptr:
		valuePointedTo, err := stringToValue(reflect.Indirect(field), parameterDesc, valueStr)
		if err != nil {
			return reflect.Value{}, err
		}
		valueDirect := valuePointedTo.Interface()
		value = &valueDirect

	case reflect.Slice:
	case reflect.Func:
	case reflect.Int:
	case reflect.Uint:
	case reflect.Interface:
	case reflect.Map:
	case reflect.UnsafePointer:
	case reflect.Invalid:
	case reflect.Uintptr:
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Array:
	case reflect.Chan:
	default:
		return reflect.Value{},
			errors.Errorf("Unsupported type '%s' for parameter '%s'", field.Type().Kind(), parameterDesc.name)
	}

	return reflect.ValueOf(value), nil
}
