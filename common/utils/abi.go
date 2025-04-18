package utils

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (u *Utils) EncodeFunctionSignature(funcName string) []byte {
	return crypto.Keccak256([]byte(funcName))[:4]
}

func (u *Utils) DecodeParameters(parameters []string, data []byte) ([]interface{}, error) {

	args := make(abi.Arguments, 0)

	for _, p := range parameters {
		arg := abi.Argument{}
		var err error
		arg.Type, err = abi.NewType(p, "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args.Unpack(data)
}

func (u *Utils) EncodeParameters(parameters []string, data []interface{}) ([]byte, error) {

	args := make(abi.Arguments, 0)

	for _, p := range parameters {
		arg := abi.Argument{}
		var err error
		arg.Type, err = abi.NewType(p, "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args.Pack(data...)
}

func (u *Utils) PackCode(signature string, args []string, params []interface{}) []byte {
	methodSig := u.EncodeFunctionSignature(signature)
	if len(args) == 0 {
		return methodSig
	}
	inputCode, err := u.EncodeParameters(args, params)
	if err != nil {
		panic(err)
	}
	code := append(methodSig, inputCode...)
	return code
}

// Equal to solidity `abi.encodePacked(args)`
func (u *Utils) AbiEncodePacked(args ...interface{}) ([]byte, error) {
	bytes := make([]byte, 0)
	for _, arg := range args {
		switch val := arg.(type) {
		case *big.Int:
			bytes = append(bytes, common.LeftPadBytes(val.Bytes(), 32)...)
		case bool:
			if val {
				bytes = append(bytes, []byte{0x0, 0x1}...)
			}
		case common.Hash:
			bytes = append(bytes, val[:]...)
		case []byte:
			bytes = append(bytes, val...)
		case common.Address:
			bytes = append(bytes, val[:]...)
		default:
			return nil, fmt.Errorf("unsupport type %T", arg)
		}
	}
	return bytes, nil
}

// ConvertArguments attempts to convert each param to the matching args type.
// Unrecognized param types are passed through unmodified.
//
// Note: The encoding/json package uses float64 for numbers by default, which is inaccurate
// for many web3 types, and unsupported here. The json.Decoder method UseNumber() will
// switch to using json.Number instead, which is accurate (full precision, backed by the
// original string) and supported here.
func (u *Utils) ConvertArguments(args abi.Arguments, params []interface{}) ([]interface{}, error) {
	if len(args) != len(params) {
		return nil, fmt.Errorf("mismatched argument (%d) and parameter (%d) counts", len(args), len(params))
	}
	var convertedParams []interface{}
	for i, input := range args {
		param, err := ConvertArgument(input.Type, params[i])
		if err != nil {
			return nil, err
		}
		convertedParams = append(convertedParams, param)
	}
	return convertedParams, nil
}

// ConvertArgument attempts to convert argument to the provided ABI type and size.
// Unrecognized types are passed through unmodified.
func ConvertArgument(abiType abi.Type, param interface{}) (interface{}, error) {
	size := abiType.Size
	// fmt.Println("INPUT TYPE:", abiType, "SIZE:", size, "Param", param)
	switch abiType.T {
	case abi.StringTy:
	case abi.BoolTy:
		if s, ok := param.(string); ok {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool %q: %v", s, err)
			}
			return val, nil
		}
	case abi.UintTy, abi.IntTy:
		if j, ok := param.(json.Number); ok {
			param = string(j)
		}
		if s, ok := param.(string); ok {
			val, ok := new(big.Int).SetString(s, 0)
			if !ok {
				return nil, fmt.Errorf("failed to parse big.Int: %s", s)
			}
			return ConvertInt(abiType.T == abi.IntTy, size, val)
		} else if i, ok := param.(*big.Int); ok {
			return ConvertInt(abiType.T == abi.IntTy, size, i)
		}
		v := reflect.ValueOf(param)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i := new(big.Int).SetInt64(v.Int())
			return ConvertInt(abiType.T == abi.IntTy, size, i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i := new(big.Int).SetUint64(v.Uint())
			return ConvertInt(abiType.T == abi.IntTy, size, i)
		case reflect.Float64, reflect.Float32:
			return nil, fmt.Errorf("floating point numbers are not valid in web3 - please use an integer or string instead (including big.Int and json.Number)")
		}
	case abi.AddressTy:
		if s, ok := param.(string); ok {
			if !common.IsHexAddress(s) {
				return nil, fmt.Errorf("invalid hex address: %s", s)
			}
			return common.HexToAddress(s), nil
		}
	case abi.SliceTy, abi.ArrayTy:
		s, ok := param.(string)
		if !ok {
			return nil, fmt.Errorf("invalid array: %s", s)
		}
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		inputArray := strings.Split(s, ",")
		switch abiType.Elem.T {

		case abi.AddressTy:
			arrayParams := make([]common.Address, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(common.Address)
			}
			return arrayParams, nil

		case abi.StringTy:
			arrayParams := make([]string, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(string)
			}
			return arrayParams, nil

		case abi.BoolTy:
			arrayParams := make([]bool, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(bool)
			}
			return arrayParams, nil

		default:
			arrayParams := make([]int, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(int)
			}
			return arrayParams, nil
		}

	case abi.BytesTy:
		if s, ok := param.(string); ok {
			val, err := hexutil.Decode(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bytes %q: %v", s, err)
			}
			return val, nil
		}
	case abi.HashTy:
		if s, ok := param.(string); ok {
			val, err := hexutil.Decode(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse hash %q: %v", s, err)
			}
			if len(val) != common.HashLength {
				return nil, fmt.Errorf("invalid hash length %d:hash must be 32 bytes", len(val))
			}
			return common.BytesToHash(val), nil
		}
	case abi.FixedBytesTy:
		switch {
		case size == 32:
			if s, ok := param.(string); ok {
				val, err := hexutil.Decode(s)
				if err != nil {
					return nil, fmt.Errorf("failed to parse hash %q: %v", s, err)
				}
				if len(val) != common.HashLength {
					return nil, fmt.Errorf("invalid hash length %d:hash must be 32 bytes", len(val))
				}
				return common.BytesToHash(val), nil
			}
		default:
			if s, ok := param.(string); ok {
				fmt.Println(s)
				val, err := hexutil.Decode(s)
				if err != nil {
					return nil, fmt.Errorf("failed to parse hash %q: %v", s, err)
				}
				if len(val) != size {
					return nil, fmt.Errorf("invalid byte array length %d: size is %d bytes", len(val), size)
				}
				arrayT := reflect.ArrayOf(size, reflect.TypeOf(byte(0)))
				array := reflect.New(arrayT).Elem()
				reflect.Copy(array, reflect.ValueOf(val))
				return array.Interface(), nil
			}
		}
	default:
		return nil, fmt.Errorf("unsupported input type %v", abiType)
	}
	return param, nil
}

// ConvertInt converts a big.Int in to the provided type.
func ConvertInt(signed bool, size int, i *big.Int) (interface{}, error) {
	if signed {
		switch {
		case size > 64:
			return i, nil
		case size > 32:
			if !i.IsInt64() {
				return nil, fmt.Errorf("integer overflows int64: %s", i)
			}
			return i.Int64(), nil
		case size > 16:
			if !i.IsInt64() || i.Int64() > math.MaxInt32 {
				return nil, fmt.Errorf("integer overflows int32: %s", i)
			}
			return int32(i.Int64()), nil
		case size > 8:
			if !i.IsInt64() || i.Int64() > math.MaxInt16 {
				return nil, fmt.Errorf("integer overflows int16: %s", i)
			}
			return int16(i.Int64()), nil
		default:
			if !i.IsInt64() || i.Int64() > math.MaxInt8 {
				return nil, fmt.Errorf("integer overflows int8: %s", i)
			}
			return int8(i.Int64()), nil
		}
	} else {
		switch {
		case size > 64:
			if i.Sign() == -1 {
				return nil, fmt.Errorf("negative value in unsigned field: %s", i)
			}
			return i, nil
		case size > 32:
			if !i.IsUint64() {
				return nil, fmt.Errorf("integer overflows uint64: %s", i)
			}
			return i.Uint64(), nil
		case size > 16:
			if !i.IsUint64() || i.Uint64() > math.MaxUint32 {
				return nil, fmt.Errorf("integer overflows uint32: %s", i)
			}
			return uint32(i.Uint64()), nil
		case size > 8:
			if !i.IsUint64() || i.Uint64() > math.MaxUint16 {
				return nil, fmt.Errorf("integer overflows uint16: %s", i)
			}
			return uint16(i.Uint64()), nil
		default:
			if !i.IsUint64() || i.Uint64() > math.MaxUint8 {
				return nil, fmt.Errorf("integer overflows uint8: %s", i)
			}
			return uint8(i.Uint64()), nil
		}
	}
}
