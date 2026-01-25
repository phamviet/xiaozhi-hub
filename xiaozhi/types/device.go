package types

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type DeviceStatus string

const DeviceCodeVerified DeviceStatus = "code_verified"
const DeviceBound DeviceStatus = "bound"

type Device struct {
	core.BaseModel
	MacAddress    string         `db:"mac_address" json:"macAddress"`
	UserId        string         `db:"user" json:"userId"`
	AgentId       string         `db:"agent" json:"agentId"`
	Status        DeviceStatus   `db:"status" json:"status"`
	Board         string         `db:"board" json:"board"`
	BindCode      string         `db:"bind_code" json:"bindCode"`
	Challenge     string         `db:"challenge" json:"challenge"`
	HmacKey       string         `db:"hmac_key" json:"hmacKey"`
	LastConnected types.DateTime `db:"last_connected" json:"lastConnected"`
}
