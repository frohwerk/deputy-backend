package images

type errUnsupportedReference error

func IsUnsupportedReference(err error) bool {
	_, ok := err.(errUnsupportedReference)
	return ok
}
