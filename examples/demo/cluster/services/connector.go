package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tutumagi/pitaya"
	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/examples/demo/protos"
)

// ConnectorRemote is a remote that will receive rpc's
type ConnectorRemote struct {
	component.Base
}

// Connector struct
type Connector struct {
	component.Base
}

// SessionData struct
type SessionData struct {
	Data map[string]interface{}
}

// Response struct
type Response struct {
	Code int32
	Msg  string
}

func reply(code int32, msg string) (*Response, error) {
	res := &Response{
		Code: code,
		Msg:  msg,
	}
	return res, nil
}

// RemoteFunc is a function that will be called remotely
func (c *ConnectorRemote) RemoteFunc(ctx context.Context, msg *protos.RPCMsg) (*protos.RPCRes, error) {
	fmt.Printf("received a remote call with this message: %s\n", msg.GetMsg())
	return &protos.RPCRes{
		Msg: msg.GetMsg(),
	}, nil
}

// Docs returns documentation
func (c *ConnectorRemote) Docs(ctx context.Context, ddd *protos.Doc) (*protos.Doc, error) {
	d, err := pitaya.Documentation(true)
	if err != nil {
		return nil, err
	}
	doc, err := json.Marshal(d)

	if err != nil {
		return nil, err
	}

	return &protos.Doc{Doc: string(doc)}, nil
}
