package data

// in-memory store
type InMemory struct {
	data []byte
}

func (r *InMemory) Set(data []byte) error {
	r.data = data
	return nil
}
func (r *InMemory) Get() ([]byte, error) {
	return r.data, nil
}

var _ Store = (*InMemory)(nil)

func NewInMemory(data []byte) *InMemory {
	return &InMemory{
		data: data,
	}
}
