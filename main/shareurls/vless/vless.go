package vless

import (
	e "XrayHelper/main/errors"
	"XrayHelper/main/serial"
	"fmt"
	"strconv"
)

const tagVless = "vless"

type VLESS struct {
	//basic
	Remarks    string
	Id         string
	Server     string
	Port       string
	Encryption string
	Flow       string
	Network    string
	Security   string

	//addon
	//ws/httpupgrade/h2->host quic->security grpc->authority
	Host string
	//ws/httpupgrade/h2->path quic->key kcp->seed grpc->serviceName
	Path string
	//tcp/kcp/quic->type grpc->mode
	Type string

	//tls
	Sni         string
	FingerPrint string
	Alpn        string
	//reality
	PublicKey string //pbk
	ShortId   string //sid
	SpiderX   string //spx
}

func (this *VLESS) GetNodeInfo() string {
	return fmt.Sprintf("Remarks: %+v, Type: VLESS, Server: %+v, Port: %+v, Flow: %+v, Network: %+v, Id: %+v", this.Remarks, this.Server, this.Port, this.Flow, this.Network, this.Id)
}

func (this *VLESS) ToOutboundWithTag(coreType string, tag string) (*serial.OrderedMap, error) {
	switch coreType {
	case "xray":
		var outboundObject serial.OrderedMap
		outboundObject.Set("mux", getMuxObjectXray(false))
		outboundObject.Set("protocol", "vless")
		outboundObject.Set("settings", getVLESSSettingsObjectXray(this))
		outboundObject.Set("streamSettings", getStreamSettingsObjectXray(this))
		outboundObject.Set("tag", tag)
		return &outboundObject, nil
	case "sing-box":
		var outboundObject serial.OrderedMap
		outboundObject.Set("type", "vless")
		outboundObject.Set("tag", tag)
		outboundObject.Set("server", this.Server)
		serverPort, _ := strconv.Atoi(this.Port)
		outboundObject.Set("server_port", serverPort)
		outboundObject.Set("uuid", this.Id)
		outboundObject.Set("flow", this.Flow)
		outboundObject.Set("tls", getVLESSTlsObjectSingbox(this))
		outboundObject.Set("transport", getVLESSTransportObjectSingbox(this))
		return &outboundObject, nil
	default:
		return nil, e.New("unsupported core type " + coreType).WithPrefix(tagVless).WithPathObj(*this)
	}
}
