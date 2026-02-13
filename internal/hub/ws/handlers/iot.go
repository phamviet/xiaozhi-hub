package handlers

import (
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

type IoTHandler struct {
	BaseHandler
}

func (h *IoTHandler) Handle(ctx Context, msg []byte) error {
	var iotMsg types.IoTMessage
	if err := h.Parse(msg, &iotMsg); err != nil {
		return err
	}

	//ctx.Logger().Debug("IoT message received", "update", iotMsg.Update)

	return nil
}
