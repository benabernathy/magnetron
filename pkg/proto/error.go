package proto

type ProtoError struct {
	Error        error
	ErrorMessage string
	Expected     any
}
