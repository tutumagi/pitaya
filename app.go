// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pitaya

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/config"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/defaultpipelines"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
	mods "github.com/tutumagi/pitaya/modules"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/serialize/json"

	"github.com/tutumagi/pitaya/session"
	"github.com/tutumagi/pitaya/timer"
	"github.com/tutumagi/pitaya/worker"
)

// ServerMode represents a server mode
type ServerMode byte

const (
	_ ServerMode = iota
	// Cluster represents a server running with connection to other servers
	Cluster
	// Standalone represents a server running without connection to other servers
	Standalone
)

// App is the base app struct
type App struct {
	acceptors        []acceptor.Acceptor
	config           *config.Config
	configured       bool
	debug            bool
	dieChan          chan bool
	heartbeat        time.Duration
	onSessionBind    func(*session.Session)
	messageEncoder   message.Encoder
	packetDecoder    codec.PacketDecoder
	packetEncoder    codec.PacketEncoder
	router           *router.Router
	rpcClient        cluster.RPCClient
	rpcServer        cluster.RPCServer
	metricsReporters []metrics.Reporter
	running          bool
	serializer       serialize.Serializer
	server           *cluster.Server
	serverMode       ServerMode
	serviceDiscovery cluster.ServiceDiscovery
	startAt          time.Time
	worker           *worker.Worker

	rootActor *actor.PID
}

// Started -> ...
// Stop(system) -> stoppingMessage(user) ->  stateStopped(user)
// 										 ->  restarting(user)
func (a *App) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logger.Log.Info("App Starting, initialize actor here")
	case *actor.Stopping:
		logger.Log.Info("App Stopping, actor is about to shut down")
	case *actor.Stopped:
		logger.Log.Info("App Stopped, actor and its children are stopped")
	case *actor.Restarting:
		logger.Log.Info("App Restarting, actor is about to restart")
	case *actor.ReceiveTimeout:
		logger.Log.Info("App ReceiveTimeout: %v", ctx.Self().String())
	case startServeMessage:
		logger.Log.Info("App server start")
		periodicMetrics()

		// 开始连接层的事件循环
		listen(ctx)
	default:
		logger.Log.Errorf("unknown message %v", msg)
	}
}

func (a *App) HandleNewConn(ctx actor.Context, conn acceptor.PlayerConn) {
	// create a client agent and startup write goroutine
	// agent := agent.NewAgent(conn,
	// 	app.packetDecoder,
	// 	app.packetEncoder,
	// 	app.serializer,
	// 	app.heartbeat,
	// 	app.config.GetInt("pitaya.buffer.agent.messages"),
	// 	app.dieChan,
	// 	app.messageEncoder,
	// 	app.metricsReporters,
	// )
	// agent.InitActor(ctx)

	// ctx.Send( actor.Started)
	// agent.Start()

	handlerService.Handle(conn)
}

type startServeMessage struct{}

var (
	app = &App{
		server:           cluster.NewServer(uuid.New().String(), "game", true, map[string]string{}),
		debug:            false,
		startAt:          time.Now(),
		dieChan:          make(chan bool),
		acceptors:        []acceptor.Acceptor{},
		packetDecoder:    codec.NewPomeloPacketDecoder(),
		packetEncoder:    codec.NewPomeloPacketEncoder(),
		metricsReporters: make([]metrics.Reporter, 0),
		serverMode:       Standalone,
		serializer:       json.NewSerializer(),
		configured:       false,
		running:          false,
		router:           router.New(),
	}

	// remoteService  *callpart.RemoteService
	handlerService *AppMsgProcessor
)

// Configure configures the app
func Configure(
	isFrontend bool,
	serverType string,
	serverMode ServerMode,
	serverMetadata map[string]string,
	cfgs ...*viper.Viper,
) {
	if app.configured {
		logger.Log.Warn("pitaya configured twice!")
	}
	app.config = config.NewConfig(cfgs...)
	if app.heartbeat == time.Duration(0) {
		app.heartbeat = app.config.GetDuration("pitaya.heartbeat.interval")
	}

	logger.Log.Debugf("begin heartbeat interval:%d sec timeout:%d sec",
		app.heartbeat/time.Second,
		2*app.heartbeat/time.Second)

	app.server.Frontend = isFrontend
	app.server.Type = serverType
	app.serverMode = serverMode
	app.server.Metadata = serverMetadata
	app.messageEncoder = message.NewMessagesEncoder(app.config.GetBool("pitaya.handler.messages.compression"))
	configureMetrics(serverType)
	configureDefaultPipelines(app.config)
	app.configured = true
}

func configureMetrics(serverType string) {
	app.metricsReporters = make([]metrics.Reporter, 0)
	constTags := app.config.GetStringMapString("pitaya.metrics.constTags")

	if app.config.GetBool("pitaya.metrics.prometheus.enabled") {
		port := app.config.GetInt("pitaya.metrics.prometheus.port")
		logger.Log.Infof("prometheus is enabled, configuring reporter on port %d", port)
		prometheus, err := metrics.GetPrometheusReporter(serverType, app.config, constTags)
		if err != nil {
			logger.Log.Errorf("failed to start prometheus metrics reporter, skipping %v", err)
		} else {
			AddMetricsReporter(prometheus)
		}
	} else {
		logger.Log.Info("prometheus is disabled, reporter will not be enabled")
	}

	if app.config.GetBool("pitaya.metrics.statsd.enabled") {
		logger.Log.Infof(
			"statsd is enabled, configuring the metrics reporter with host: %s",
			app.config.Get("pitaya.metrics.statsd.host"),
		)
		metricsReporter, err := metrics.NewStatsdReporter(
			app.config,
			serverType,
			constTags,
		)
		if err != nil {
			logger.Log.Errorf("failed to start statds metrics reporter, skipping %v", err)
		} else {
			logger.Log.Info("successfully configured statsd metrics reporter")
			AddMetricsReporter(metricsReporter)
		}
	}
}

func configureDefaultPipelines(config *config.Config) {
	if config.GetBool("pitaya.defaultpipelines.structvalidation.enabled") {
		BeforeHandler(defaultpipelines.StructValidatorInstance.Validate)
	}
}

func startDefaultSD() {
	// initialize default service discovery
	var err error
	app.serviceDiscovery, err = cluster.NewEtcdServiceDiscovery(
		app.config,
		app.server,
		app.dieChan,
	)
	if err != nil {
		logger.Log.Fatalf("error starting cluster service discovery component: %s", err.Error())
	}
}

func startDefaultRPCServer() {
	// initialize default rpc server
	rpcServer, err := cluster.NewNatsRPCServer(app.config, app.server, app.metricsReporters, app.dieChan)
	if err != nil {
		logger.Log.Fatalf("error starting cluster rpc server component: %s", err.Error())
	}
	SetRPCServer(rpcServer)
}

func startDefaultRPCClient() {
	// initialize default rpc client
	rpcClient, err := cluster.NewNatsRPCClient(app.config, app.server, app.metricsReporters, app.dieChan)
	if err != nil {
		logger.Log.Fatalf("error starting cluster rpc client component: %s", err.Error())
	}
	SetRPCClient(rpcClient)
}

func initSysRemotes() {
	// TODO sys应该作为一个 service 创建
	// 	sys := &remote.Sys{}
	// 	RegisterRemote(sys,
	// 		component.WithName("sys"),
	// 		component.WithNameFunc(strings.ToLower),
	// 	)
}

func periodicMetrics() {
	period := app.config.GetDuration("pitaya.metrics.periodicMetrics.period")
	go metrics.ReportSysMetrics(app.metricsReporters, period)

	if app.worker.Started() {
		go worker.Report(app.metricsReporters, period)
	}
}

// Start starts the app
func Start() {
	if !app.configured {
		logger.Log.Fatal("starting app without configuring it first! call pitaya.Configure()")
	}

	if !app.server.Frontend && len(app.acceptors) > 0 {
		logger.Log.Fatal("acceptors are not allowed on backend servers")
	}

	if app.server.Frontend && len(app.acceptors) == 0 {
		logger.Log.Fatal("frontend servers should have at least one configured acceptor")
	}

	// if app.serverMode == Cluster {
	if app.serviceDiscovery == nil {
		logger.Log.Warn("creating default service discovery because cluster mode is enabled, " +
			"if you want to specify yours, use pitaya.SetServiceDiscoveryClient")
		startDefaultSD()
	}
	if app.rpcServer == nil {
		logger.Log.Warn("creating default rpc server because cluster mode is enabled, " +
			"if you want to specify yours, use pitaya.SetRPCServer")
		startDefaultRPCServer()
	}
	if app.rpcClient == nil {
		logger.Log.Warn("creating default rpc client because cluster mode is enabled, " +
			"if you want to specify yours, use pitaya.SetRPCClient")
		startDefaultRPCClient()
	}

	// by tufei 暂不使用grpc
	// if reflect.TypeOf(app.rpcClient) == reflect.TypeOf(&cluster.GRPCClient{}) {
	// 	app.serviceDiscovery.AddListener(app.rpcClient.(*cluster.GRPCClient))
	// }

	if err := RegisterModuleBefore(app.rpcServer, "rpcServer"); err != nil {
		logger.Log.Fatal("failed to register rpc server module: %s", err.Error())
	}
	if err := RegisterModuleBefore(app.rpcClient, "rpcClient"); err != nil {
		logger.Log.Fatal("failed to register rpc client module: %s", err.Error())
	}
	// set the service discovery as the last module to be started to ensure
	// all modules have been properly initialized before the server starts
	// receiving requests from other pitaya servers
	if err := RegisterModuleAfter(app.serviceDiscovery, "serviceDiscovery"); err != nil {
		logger.Log.Fatal("failed to register service discovery module: %s", err.Error())
	}

	app.router.SetServiceDiscovery(app.serviceDiscovery)

	// }

	sys := actor.NewActorSystem()

	handlerService = NewAppProcessor(
		app.dieChan,
		app.packetDecoder,
		app.packetEncoder,
		app.serializer,
		app.heartbeat,
		app.config.GetInt("pitaya.buffer.agent.messages"),
		app.config.GetInt("pitaya.buffer.handler.localprocess"),
		app.config.GetInt("pitaya.buffer.handler.remoteprocess"),

		app.server,
		app.messageEncoder,
		app.metricsReporters,
		app.rpcClient,
		app.rpcServer,
		app.serviceDiscovery,
		app.router,
		sys,
	)

	// remoteService = callpart.NewRemoteService(
	// 	app.rpcClient,
	// 	app.rpcServer,
	// 	app.serviceDiscovery,
	// 	app.packetEncoder,
	// 	app.serializer,
	// 	app.router,
	// 	app.messageEncoder,
	// 	app.server,
	// )

	app.rpcServer.SetPitayaServer(handlerService)

	initSysRemotes()

	rootProps := actor.PropsFromProducer(func() actor.Actor {
		return app
	})

	rootCtx := sys.Root
	app.rootActor, _ = rootCtx.SpawnNamed(rootProps, "__root__")

	rootCtx.Send(app.rootActor, startServeMessage{})

	defer func() {
		timer.GlobalTicker.Stop()
		app.running = false
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case <-app.dieChan:
		logger.Log.Warn("the app will shutdown in a few seconds")
	case s := <-sg:
		logger.Log.Warn("got signal: ", s, ", shutting down...")
		close(app.dieChan)
	}

	logger.Log.Warn("server is stopping...")

	session.CloseAll()
	shutdownModules()
	// shutdownComponents()
}

func listen(ctx actor.Context) {
	// startupComponents()
	// create global ticker instance, timer precision could be customized
	// by SetTimerPrecision
	timer.GlobalTicker = time.NewTicker(timer.Precision)

	logger.Log.Infof("starting server %s:%s", app.server.Type, app.server.ID)
	for i := 0; i < app.config.GetInt("pitaya.concurrency.handler.dispatch"); i++ {
		go handlerService.Dispatch(i)
	}
	for _, acc := range app.acceptors {
		a := acc
		// 处理新连接的消息
		go func() {
			for conn := range a.GetConnChan() {
				go app.HandleNewConn(ctx, conn)
			}
		}()

		// 监听并处理新连接
		go func() {
			a.ListenAndServe()
		}()

		logger.Log.Infof("listening with acceptor %s on addr %s", reflect.TypeOf(a), a.GetAddr())
	}

	if app.serverMode == Cluster && app.server.Frontend && app.config.GetBool("pitaya.session.unique") {
		unique := mods.NewUniqueSession(app.server, app.rpcServer, app.rpcClient)
		handlerService.AddRemoteBindingListener(unique)
		RegisterModule(unique, "uniqueSession")
	}

	startModules()

	logger.Log.Info("all modules started!")

	app.running = true
}
