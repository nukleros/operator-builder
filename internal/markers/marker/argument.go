package marker

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/parser"
)

var (
	ErrWrongType = errors.New("incorrect type")
	ErrUnmarshal = errors.New("unable to unmarshal arg value")
)

// Argument is the type of a marker argument.
type Argument struct {
	Name      string
	FieldName string
	// Type is the type of this argument.
	Type reflect.Type
	// Optional indicates if this argument is optional.
	Optional bool
	// Pointer indicates if this argument was a pointer (this is really only
	// needed for deserialization, and should alway imply optional).
	Pointer bool

	Value reflect.Value

	isSet bool
}

func (a Argument) String() string {
	if a.Optional {
		return fmt.Sprintf("<optional arg %s>", a.Type)
	}

	return fmt.Sprintf("<arg %s>", a.Type)
}

func ArgumentFromField(field *reflect.StructField) (Argument, error) {
	arg := Argument{
		Name:      lowerCamelCase(field.Name),
		FieldName: field.Name,
		Type:      field.Type,
		Optional:  false,
	}

	if tag, found := field.Tag.Lookup("marker"); found {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			arg.Name = parts[0]
		}

		for _, opt := range parts[1:] {
			if opt == "optional" {
				arg.Optional = true
			}
		}
	}

	if err := arg.SetTypeInfo(); err != nil {
		return arg, err
	}

	return arg, nil
}

// ArgumentFromType constructs an Argument by examining the given reflect.Type.
func (a *Argument) SetTypeInfo() error {
	// interfaceType is a pre-computed reflect.Type representing the empty interface.
	interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()

	if a.Type == interfaceType {
		a.Value = reflect.Indirect(reflect.New(a.Type))
	}

	if a.Type.Kind() == reflect.Ptr {
		a.Pointer = true
		a.Optional = true
	}

	return nil
}

func (a *Argument) SetValue(value interface{}) error {
	unmarshalerType := reflect.TypeOf((*parser.Unmarshaler)(nil)).Elem()

	a.InitializeValue()

	switch {
	case reflect.PtrTo(a.Type).Implements(unmarshalerType):
		if s, ok := value.(string); ok {
			if errs := a.Value.Addr().MethodByName("UnmarshalMarkerArg").Call([]reflect.Value{reflect.ValueOf(s)}); errs[0].Interface() != nil {
				return fmt.Errorf("%w %q, %s", ErrUnmarshal, value, errs[0])
			}

			a.isSet = true

			return nil
		}

		return fmt.Errorf("%w, cannot convert %v to string", ErrUnmarshal, value)
	case a.Pointer:
		if !reflect.TypeOf(value).ConvertibleTo(a.Type.Elem()) {
			return fmt.Errorf("%w, wanted %q but received %q", ErrWrongType, a.Type.Elem(), reflect.TypeOf(value))
		}

		a.Value.Elem().Set(reflect.ValueOf(value))

		a.isSet = true

		return nil
	case !reflect.TypeOf(value).ConvertibleTo(a.Type):
		return fmt.Errorf("%w, wanted %q but received %q", ErrWrongType, a.Type, reflect.TypeOf(value))
	default:
		a.Value.Set(reflect.ValueOf(value))
		a.isSet = true

		return nil
	}
}

func (a *Argument) InitializeValue() {
	if a.Pointer {
		a.Value = reflect.New(a.Type.Elem())

		return
	}

	a.Value = reflect.Indirect(reflect.New(a.Type))
}
