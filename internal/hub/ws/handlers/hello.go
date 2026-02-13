package handlers

type HelloHandler struct {
	BaseHandler
}

func (h *HelloHandler) Handle(ctx Context, msg []byte) error {
	return nil
}
