package essos

type Operation interface {
	Create(interface{}) string

	Read() string

	Update() string

	Delete() string

	Do() string
}
