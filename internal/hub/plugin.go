package hub

type Plugin interface {
	Name() string
	Initialize(hub *Hub) error
}
