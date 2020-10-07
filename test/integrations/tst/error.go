package tst

const (
	ErrNotFound   Error = "not_found"
	ErrOutOfLimit Error = "out_of_limit"
)

type Error string

func (o Error) Error() string {
	return string(o)
}
