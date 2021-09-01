package parser

type Unmarshaler interface {
	UnmarshalMarkerArg(in string) error
}
