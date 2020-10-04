package kvs

type KVS interface {
	Add(key, value string) error
	Get(key string) (string, error)
}

type Client struct {
}

func NewClient() KVS {
	return &Client{}
}

func (o *Client) Add(key, value string) error {
	return nil
}

func (o *Client) Get(key string) (string, error) {
	return "", nil
}
