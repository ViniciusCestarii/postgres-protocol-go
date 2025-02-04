package types

import (
	"database/sql/driver"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
)

var (
	driverValuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	appenderType     = reflect.TypeOf((*ValueAppender)(nil)).Elem()
)

type AppenderFunc func([]byte, reflect.Value, int) []byte

var appenders []AppenderFunc

// nolint
func init() {
	appenders = []AppenderFunc{
		reflect.Bool:          appendBoolValue,
		reflect.Int:           appendIntValue,
		reflect.Int8:          appendIntValue,
		reflect.Int16:         appendIntValue,
		reflect.Int32:         appendIntValue,
		reflect.Int64:         appendIntValue,
		reflect.Uint:          appendUintValue,
		reflect.Uint8:         appendUintValue,
		reflect.Uint16:        appendUintValue,
		reflect.Uint32:        appendUintValue,
		reflect.Uint64:        appendUintValue,
		reflect.Uintptr:       nil,
		reflect.Float32:       appendFloat32Value,
		reflect.Float64:       appendFloat64Value,
		reflect.Complex64:     nil,
		reflect.Complex128:    nil,
		reflect.Chan:          nil,
		reflect.Func:          nil,
		reflect.Interface:     appendIfaceValue,
		reflect.Ptr:           nil,
		reflect.String:        appendStringValue,
		reflect.UnsafePointer: nil,
	}
}

var appendersMap sync.Map

// RegisterAppender registers an appender func for the value type.
// Expecting to be used only during initialization, it panics
// if there is already a registered appender for the given type.
func RegisterAppender(value interface{}, fn AppenderFunc) {
	registerAppender(reflect.TypeOf(value), fn)
}

func registerAppender(typ reflect.Type, fn AppenderFunc) {
	_, loaded := appendersMap.LoadOrStore(typ, fn)
	if loaded {
		err := fmt.Errorf("pg: appender for the type=%s is already registered",
			typ.String())
		panic(err)
	}
}

func Appender(typ reflect.Type) AppenderFunc {
	if v, ok := appendersMap.Load(typ); ok {
		return v.(AppenderFunc)
	}
	fn := appender(typ)
	_, _ = appendersMap.LoadOrStore(typ, fn)
	return fn
}

func appender(typ reflect.Type) AppenderFunc {
	if typ.Implements(appenderType) {
		return appendAppenderValue
	}
	if typ.Implements(driverValuerType) {
		return appendDriverValuerValue
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Ptr:
		return ptrAppenderFunc(typ)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return appendBytesValue
		}
	}
	return appenders[kind]
}

func ptrAppenderFunc(typ reflect.Type) AppenderFunc {
	appender := Appender(typ.Elem())
	return func(b []byte, v reflect.Value, flags int) []byte {
		if v.IsNil() {
			return AppendNull(b, flags)
		}
		return appender(b, v.Elem(), flags)
	}
}

func appendValue(b []byte, v reflect.Value, flags int) []byte {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return AppendNull(b, flags)
	}
	appender := Appender(v.Type())
	return appender(b, v, flags)
}

func appendIfaceValue(b []byte, v reflect.Value, flags int) []byte {
	return Append(b, v.Interface(), flags)
}

func appendBoolValue(b []byte, v reflect.Value, _ int) []byte {
	return appendBool(b, v.Bool())
}

func appendIntValue(b []byte, v reflect.Value, _ int) []byte {
	return strconv.AppendInt(b, v.Int(), 10)
}

func appendUintValue(b []byte, v reflect.Value, _ int) []byte {
	return strconv.AppendUint(b, v.Uint(), 10)
}

func appendFloat32Value(b []byte, v reflect.Value, flags int) []byte {
	return appendFloat(b, v.Float(), flags, 32)
}

func appendFloat64Value(b []byte, v reflect.Value, flags int) []byte {
	return appendFloat(b, v.Float(), flags, 64)
}

func appendBytesValue(b []byte, v reflect.Value, flags int) []byte {
	return AppendBytes(b, v.Bytes(), flags)
}

// func appendArrayBytesValue(b []byte, v reflect.Value, flags int) []byte {
// 	if v.CanAddr() {
// 		return AppendBytes(b, v.Slice(0, v.Len()).Bytes(), flags)
// 	}

// 	buf := bufpool.Get(v.Len())

// 	tmp := buf.Bytes()
// 	reflect.Copy(reflect.ValueOf(tmp), v)
// 	b = AppendBytes(b, tmp, flags)

// 	bufpool.Put(buf)

// 	return b
// }

func appendStringValue(b []byte, v reflect.Value, flags int) []byte {
	return AppendString(b, v.String(), flags)
}

// func appendTimeValue(b []byte, v reflect.Value, flags int) []byte {
// 	tm := v.Interface().(time.Time)
// 	return AppendTime(b, tm, flags)
// }

func appendIPValue(b []byte, v reflect.Value, flags int) []byte {
	ip := v.Interface().(net.IP)
	return AppendString(b, ip.String(), flags)
}

func appendIPNetValue(b []byte, v reflect.Value, flags int) []byte {
	ipnet := v.Interface().(net.IPNet)
	return AppendString(b, ipnet.String(), flags)
}

func appendAppenderValue(b []byte, v reflect.Value, flags int) []byte {
	return appendAppender(b, v.Interface().(ValueAppender), flags)
}

func appendDriverValuerValue(b []byte, v reflect.Value, flags int) []byte {
	return appendDriverValuer(b, v.Interface().(driver.Valuer), flags)
}

func appendDriverValuer(b []byte, v driver.Valuer, flags int) []byte {
	value, err := v.Value()
	if err != nil {
		return AppendError(b, err)
	}
	return Append(b, value, flags)
}
