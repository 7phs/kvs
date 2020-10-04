package storages

const (
	ErrOutOfLimit  Error = "out_of_the_limit"
	ErrKeyNotFound Error = "key_not_found"
)

type Error string

func (o Error) Error() string {
	return string(o)
}
