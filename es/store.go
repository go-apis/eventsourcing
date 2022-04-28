package es

type Store interface{}

func NewStore(url string) (Store, error) {
	// todo support different types of stores

	return NewDbStore(url)
}
