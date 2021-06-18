package app

import (
	"context"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/config"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/constants"
	pcontext "github.com/tutumagi/pitaya/context"
	"github.com/tutumagi/pitaya/docgenerator"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/session"
	"github.com/tutumagi/pitaya/tracing"
	"github.com/tutumagi/pitaya/worker"
)

// AddAcceptor adds a new acceptor to app
func AddAcceptor(ac acceptor.Acceptor) {
	if !app.server.Frontend {
		logger.Log.Error("tried to add an acceptor to a backend server, skipping")
		return
	}
	app.acceptors = append(app.acceptors, ac)
}

// GetDieChan gets the channel that the app sinalizes when its going to die
func GetDieChan() chan bool {
	return app.dieChan
}

// SetDebug toggles debug on/off
func SetDebug(debug bool) {
	app.debug = debug
}

// SetPacketDecoder changes the decoder used to parse messages received
func SetPacketDecoder(d codec.PacketDecoder) {
	app.packetDecoder = d
}

// SetPacketEncoder changes the encoder used to package outgoing messages
func SetPacketEncoder(e codec.PacketEncoder) {
	app.packetEncoder = e
}

// SetHeartbeatTime sets the heartbeat time
func SetHeartbeatTime(interval time.Duration) {
	app.heartbeat = interval
}

// SetLogger logger setter
func SetLogger(l logger.Logger) {
	logger.Log = l
}

// GetServerID returns the generated server id
func GetServerID() string {
	return app.server.ID
}

// GetConfig gets the pitaya config instance
func GetConfig() *config.Config {
	return app.config
}

// GetMetricsReporters gets registered metrics reporters
func GetMetricsReporters() []metrics.Reporter {
	return app.metricsReporters
}

// SetRPCServer to be used
func SetRPCServer(s cluster.RPCServer) {
	app.rpcServer = s
}

// SetRPCClient to be used
func SetRPCClient(s cluster.RPCClient) {
	app.rpcClient = s
}

// SetServiceDiscoveryClient to be used
func SetServiceDiscoveryClient(s cluster.ServiceDiscovery) {
	app.serviceDiscovery = s
}

// SetSerializer customize application serializer, which automatically Marshal
// and UnMarshal handler payload
func SetSerializer(seri serialize.Serializer) {
	app.serializer = seri
}

// GetSerializer gets the app serializer
func GetSerializer() serialize.Serializer {
	return app.serializer
}

// GetServer gets the local server instance
func GetServer() *cluster.Server {
	return app.server
}

// GetServerByID returns the server with the specified id
func GetServerByID(id string) (*cluster.Server, error) {
	return app.serviceDiscovery.GetServer(id)
}

// GetServersByType get all servers of type
func GetServersByType(t string) (map[string]*cluster.Server, error) {
	return app.serviceDiscovery.GetServersByType(t)
}

// GetServers get all servers
func GetServers() []*cluster.Server {
	return app.serviceDiscovery.GetServers()
}

// AddMetricsReporter to be used
func AddMetricsReporter(mr metrics.Reporter) {
	app.metricsReporters = append(app.metricsReporters, mr)
}

// SetDictionary sets routes map
func SetDictionary(dict map[string]uint16) error {
	if app.running {
		return constants.ErrChangeDictionaryWhileRunning
	}
	return message.SetDictionary(dict)
}

// AddRoute adds a routing function to a server type
func AddRoute(
	serverType string,
	routingFunction router.RoutingFunc,
) error {
	if app.router != nil {
		if app.running {
			return constants.ErrChangeRouteWhileRunning
		}
		app.router.AddRoute(serverType, routingFunction)
	} else {
		return constants.ErrRouterNotInitialized
	}
	return nil
}

// Shutdown send a signal to let 'pitaya' shutdown itself.
func Shutdown() {
	select {
	case <-app.dieChan: // prevent closing closed channel
	default:
		close(app.dieChan)
	}
}

// Error creates a new error with a code, message and metadata
func Error(err error, code string, metadata ...map[string]string) *errors.Error {
	return errors.NewError(err, code, metadata...)
}

// GetSessionFromCtx retrieves a session from a given context
func GetSessionFromCtx(ctx context.Context) *session.Session {
	sessionVal := ctx.Value(constants.SessionCtxKey)
	if sessionVal == nil {
		logger.Log.Debug("ctx doesn't contain a session, are you calling GetSessionFromCtx from inside a remote?")
		return nil
	}
	return sessionVal.(*session.Session)
}

// GetDefaultLoggerFromCtx returns the default logger from the given context
func GetDefaultLoggerFromCtx(ctx context.Context) logger.Logger {
	l := ctx.Value(constants.LoggerCtxKey)
	if l == nil {
		return logger.Log
	}

	return l.(logger.Logger)
}

// AddMetricTagsToPropagateCtx adds a key and metric tags that will
// be propagated through RPC calls. Use the same tags that are at
// 'pitaya.metrics.additionalTags' config
func AddMetricTagsToPropagateCtx(
	ctx context.Context,
	tags map[string]string,
) context.Context {
	return pcontext.AddToPropagateCtx(ctx, constants.MetricTagsKey, tags)
}

// AddToPropagateCtx adds a key and value that will be propagated through RPC calls
func AddToPropagateCtx(ctx context.Context, key string, val interface{}) context.Context {
	return pcontext.AddToPropagateCtx(ctx, key, val)
}

// GetFromPropagateCtx adds a key and value that came through RPC calls
func GetFromPropagateCtx(ctx context.Context, key string) interface{} {
	return pcontext.GetFromPropagateCtx(ctx, key)
}

// ExtractSpan retrieves an opentracing span context from the given context
// The span context can be received directly or via an RPC call
func ExtractSpan(ctx context.Context) (opentracing.SpanContext, error) {
	return tracing.ExtractSpan(ctx)
}

// Documentation returns handler and remotes documentacion
func Documentation(getPtrNames bool) (map[string]interface{}, error) {

	return metapart.Documents(getPtrNames)
	// handlerDocs, err := handlerService.DocsHandler(getPtrNames)
	// if err != nil {
	// 	return nil, err
	// }
	// remoteDocs, err := handlerService.DocsRemote(getPtrNames)
	// if err != nil {
	// 	return nil, err
	// }
	// return map[string]interface{}{
	// 	"handlers": handlerDocs,
	// 	"remotes":  remoteDocs,
	// }, nil
}

// AddGRPCInfoToMetadata adds host, external host and
// port into metadata
func AddGRPCInfoToMetadata(
	metadata map[string]string,
	region string,
	host, port string,
	externalHost, externalPort string,
) map[string]string {
	metadata[constants.GRPCHostKey] = host
	metadata[constants.GRPCPortKey] = port
	metadata[constants.GRPCExternalHostKey] = externalHost
	metadata[constants.GRPCExternalPortKey] = externalPort
	metadata[constants.RegionKey] = region
	return metadata
}

// Descriptor returns the protobuf message descriptor for a given message name
func Descriptor(protoName string) ([]byte, error) {
	return docgenerator.ProtoDescriptors(protoName)
}

// StartWorker configures, starts and returns pitaya worker
func StartWorker(config *config.Config) error {
	var err error
	app.worker, err = worker.NewWorker(config)
	if err != nil {
		return err
	}

	app.worker.Start()

	return nil
}

// RegisterRPCJob registers rpc job to execute jobs with retries
func RegisterRPCJob(rpcJob worker.RPCJob) error {
	err := app.worker.RegisterRPCJob(rpcJob)
	return err
}
