package basepart

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/tutumagi/pitaya/engine/common"
)

func (e *Entity) CallService(ctx context.Context, serviceName string, routeStr string, reply proto.Message, arg proto.Message) error {
	entityID := common.ServiceID(serviceName)
	entityType := common.ServiceTypeName(serviceName)

	return caller.call(ctx, "", entityID, entityType, routeStr, reply, arg)
}

func (e *Entity) SendService(ctx context.Context, serviceName string, routeStr string, arg proto.Message) error {
	entityID := common.ServiceID(serviceName)
	entityType := common.ServiceTypeName(serviceName)

	return caller.call(ctx, "", entityID, entityType, routeStr, nil, arg)
}

func (e *Entity) CallServiceTo(
	ctx context.Context,
	serverID string,
	serviceName string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	entityID := common.ServiceID(serviceName)
	entityType := common.ServiceTypeName(serviceName)

	return caller.call(ctx, serverID, entityID, entityType, routeStr, reply, arg)
}

func (e *Entity) SendServiceTo(
	ctx context.Context,
	serverID string,
	serviceName string,
	routeStr string,
	arg proto.Message,
) error {
	entityID := common.ServiceID(serviceName)
	entityType := common.ServiceTypeName(serviceName)

	return caller.call(ctx, serverID, entityID, entityType, routeStr, nil, arg)
}

func (e *Entity) CallEntity(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	return caller.call(ctx, "", entityID, entityType, routeStr, reply, arg)
}

func (e *Entity) CallEntityTo(
	ctx context.Context,
	serverID string,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	return caller.call(ctx, serverID, entityID, entityType, routeStr, reply, arg)
}

func (e *Entity) SendEntity(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	arg proto.Message,
) error {
	return caller.call(ctx, "", entityID, entityType, routeStr, nil, arg)
}

func (e *Entity) SendEntityTo(
	ctx context.Context,
	serverID string,
	entityID,
	entityType string,
	routeStr string,
	arg proto.Message,
) error {
	return caller.call(ctx, serverID, entityID, entityType, routeStr, nil, arg)
}
