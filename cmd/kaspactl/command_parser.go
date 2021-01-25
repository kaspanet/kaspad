package main

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
)

func parseCommand(args []string, requestDescs []*requestDescription) (*protowire.KaspadMessage, error) {
	commandName, parameters := args[0], args[1:]

	var requestDesc *requestDescription
	for _, rd := range requestDescs {
		if rd.name == commandName {
			requestDesc = rd
			break
		}
	}
	if requestDesc == nil {
		return nil, errors.Errorf("unknown command: %s. Use --list-commands to list all commands", commandName)
	}
	if len(parameters) != len(requestDesc.parameters) {
		return nil, errors.Errorf("command '%s' expects %d parameters but got %d",
			commandName, len(requestDesc.parameters), len(parameters))
	}

	request := reflect.New(unwrapRequestType(requestDesc.typeof))
	for i, parameterDesc := range requestDesc.parameters {
		field := request.Elem().FieldByName(parameterDesc.name)
		parameter := parameters[i]
		err := setField(field, parameterDesc, parameter)
		if err != nil {
			return nil, err
		}
	}

	requestWrapper := reflect.New(requestDesc.typeof)
	requestWrapper.Elem().Field(0).Set(request)

	kaspadMessage := reflect.New(reflect.TypeOf(protowire.KaspadMessage{}))
	kaspadMessage.Elem().FieldByName("Payload").Set(requestWrapper)
	return kaspadMessage.Interface().(*protowire.KaspadMessage), nil
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
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.Int8:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 8)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = int8(valueInt64)
	case reflect.Int16:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 16)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = int16(valueInt64)
	case reflect.Int32:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = int32(valueInt64)
	case reflect.Int64:
		value, err = strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.Uint8:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 8)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = uint8(valueUInt64)
	case reflect.Uint16:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 16)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = uint16(valueUInt64)
	case reflect.Uint32:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 32)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = uint32(valueUInt64)
	case reflect.Uint64:
		value, err = strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.Float32:
		var valueFloat64 float64
		valueFloat64, err = strconv.ParseFloat(valueStr, 32)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = float32(valueFloat64)
	case reflect.Float64:
		value, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.String:
		value = valueStr
	case reflect.Struct:
		pointer := reflect.New(field.Type()) // create pointer to this type
		fieldInterface := pointer.Interface().(proto.Message)
		err := protojson.Unmarshal([]byte(valueStr), fieldInterface)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = fieldInterface.ProtoReflect().Interface()
	case reflect.Ptr:
		valuePointedTo, err := stringToValue(reflect.New(field.Type().Elem()).Elem(), parameterDesc, valueStr)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
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
