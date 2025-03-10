package protostruct

//https://gist.github.com/jsmouret/2bc876e8def6c63410556350eca3e43d
import (
	"fmt"
	"reflect"

	st "github.com/golang/protobuf/ptypes/struct"
)

// ToStruct converts a map[string]interface{} to a ptypes.Struct
func ToStruct(v map[string]interface{}) *st.Struct {
	size := len(v)
	if size == 0 {
		return nil

	}
	fields := make(map[string]*st.Value, size)
	for k, v := range v {
		fields[k] = ToValue(v)

	}
	return &st.Struct{
		Fields: fields,
	}

}

// ToValue converts an interface{} to a ptypes.Value
func ToValue(v interface{}) *st.Value {
	switch v := v.(type) {
	case nil:
		return nil
	case bool:
		return &st.Value{
			Kind: &st.Value_BoolValue{
				BoolValue: v,
			},
		}
	case int:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int8:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int32:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint8:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint32:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float32:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: v,
			},
		}
	case string:
		return &st.Value{
			Kind: &st.Value_StringValue{
				StringValue: v,
			},
		}
	case error:
		return &st.Value{
			Kind: &st.Value_StringValue{
				StringValue: v.Error(),
			},
		}
	default:
		// Fallback to reflection for other types
		return toValue(reflect.ValueOf(v))

	}

}

func toValue(v reflect.Value) *st.Value {
	switch v.Kind() {
	case reflect.Bool:
		return &st.Value{
			Kind: &st.Value_BoolValue{
				BoolValue: v.Bool(),
			},
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v.Int()),
			},
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(v.Uint()),
			},
		}
	case reflect.Float32, reflect.Float64:
		return &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: v.Float(),
			},
		}
	case reflect.Ptr:
		if v.IsNil() {
			return nil

		}
		return toValue(reflect.Indirect(v))
	case reflect.Array, reflect.Slice:
		size := v.Len()
		if size == 0 {
			return nil

		}
		values := make([]*st.Value, size)
		for i := 0; i < size; i++ {
			values[i] = toValue(v.Index(i))

		}
		return &st.Value{
			Kind: &st.Value_ListValue{
				ListValue: &st.ListValue{
					Values: values,
				},
			},
		}
	case reflect.Struct:
		t := v.Type()
		size := v.NumField()
		if size == 0 {
			return nil

		}
		fields := make(map[string]*st.Value, size)
		for i := 0; i < size; i++ {
			name := t.Field(i).Name
			// Better way?
			if len(name) > 0 && 'A' <= name[0] && name[0] <= 'Z' {
				fields[name] = toValue(v.Field(i))

			}

		}
		if len(fields) == 0 {
			return nil

		}
		return &st.Value{
			Kind: &st.Value_StructValue{
				StructValue: &st.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Map:
		keys := v.MapKeys()
		if len(keys) == 0 {
			return nil

		}
		fields := make(map[string]*st.Value, len(keys))
		for _, k := range keys {
			if k.Kind() == reflect.String {
				fields[k.String()] = toValue(v.MapIndex(k))

			}

		}
		if len(fields) == 0 {
			return nil

		}
		return &st.Value{
			Kind: &st.Value_StructValue{
				StructValue: &st.Struct{
					Fields: fields,
				},
			},
		}
	default:
		// Last resort
		return &st.Value{
			Kind: &st.Value_StringValue{
				StringValue: fmt.Sprint(v),
			},
		}

	}

}
