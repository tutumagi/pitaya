package common

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type CallImper interface {
	Call(
		ctx context.Context,
		serverID string,
		entityID,
		entityType string,
		routeStr string,
		reply proto.Message,
		arg proto.Message,
	) error
}

type Caller struct {
	imp CallImper
}

func NewCaller(imp CallImper) *Caller {
	return &Caller{
		imp: imp,
	}
}

func (c *Caller) CallService(ctx context.Context, serviceName string, routeStr string, reply proto.Message, arg proto.Message) error {
	entityID := ServiceID(serviceName)
	entityType := ServiceTypeName(serviceName)

	return c.imp.Call(ctx, "", entityID, entityType, routeStr, reply, arg)
}

func (c *Caller) SendService(ctx context.Context, serviceName string, routeStr string, arg proto.Message) error {
	entityID := ServiceID(serviceName)
	entityType := ServiceTypeName(serviceName)

	return c.imp.Call(ctx, "", entityID, entityType, routeStr, nil, arg)
}

func (c *Caller) CallServiceTo(
	ctx context.Context,
	serverID string,
	serviceName string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	entityID := ServiceID(serviceName)
	entityType := ServiceTypeName(serviceName)

	return c.imp.Call(ctx, serverID, entityID, entityType, routeStr, reply, arg)
}

func (c *Caller) SendServiceTo(
	ctx context.Context,
	serverID string,
	serviceName string,
	routeStr string,
	arg proto.Message,
) error {
	entityID := ServiceID(serviceName)
	entityType := ServiceTypeName(serviceName)

	return c.imp.Call(ctx, serverID, entityID, entityType, routeStr, nil, arg)
}

func (c *Caller) CallEntity(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	return c.imp.Call(ctx, "", entityID, entityType, routeStr, reply, arg)
}

func (c *Caller) CallEntityTo(
	ctx context.Context,
	serverID string,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	return c.imp.Call(ctx, serverID, entityID, entityType, routeStr, reply, arg)
}

func (c *Caller) SendEntity(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	arg proto.Message,
) error {
	return c.imp.Call(ctx, "", entityID, entityType, routeStr, nil, arg)
}

func (c *Caller) SendEntityTo(
	ctx context.Context,
	serverID string,
	entityID,
	entityType string,
	routeStr string,
	arg proto.Message,
) error {
	return c.imp.Call(ctx, serverID, entityID, entityType, routeStr, nil, arg)
}
