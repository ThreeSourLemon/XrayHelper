package switches

import (
	"XrayHelper/main/errors"
	"XrayHelper/main/switches/clash"
	"XrayHelper/main/switches/ray"
)

// Switch implement this interface, that program can deal different core config switch
type Switch interface {
	Execute(args []string) (bool, error)
}

func NewSwitch(coreType string) (Switch, error) {
	switch coreType {
	case "xray", "v2ray", "sing-box":
		return new(ray.RaySwitch), nil
	case "clash", "clash.premium", "clash.meta":
		return new(clash.ClashSwitch), nil
	default:
		return nil, errors.New("unsupported core type " + coreType).WithPrefix("switches")
	}
}
