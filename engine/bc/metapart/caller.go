package metapart

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type Caller interface {
	CallService(
		ctx context.Context,
		serviceName string,
		routeStr string,
		reply proto.Message,
		arg proto.Message,
	) error
	CallServiceTo(
		ctx context.Context,
		serverID string,
		serviceName string,
		routeStr string,
		reply proto.Message,
		arg proto.Message,
	) error

	SendService(
		ctx context.Context,
		serviceName string,
		routeStr string,
		arg proto.Message,
	) error
	SendServiceTo(
		ctx context.Context,
		serverID string,
		serviceName string,
		routeStr string,
		arg proto.Message,
	) error

	CallEntity(
		ctx context.Context,
		entityID,
		entityType string,
		routeStr string,
		reply proto.Message,
		arg proto.Message,
	) error
	CallEntityTo(
		ctx context.Context,
		serverID,
		entityID,
		entityType string,
		routeStr string,
		reply proto.Message,
		arg proto.Message,
	) error
	SendEntity(
		ctx context.Context,
		entityID,
		entityType string,
		routeStr string,
		arg proto.Message,
	) error
	SendEntityTo(
		ctx context.Context,
		serverID string,
		entityID,
		entityType string,
		routeStr string,
		arg proto.Message,
	) error
}
