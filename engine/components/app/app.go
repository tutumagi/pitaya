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

package app

import (
	"os"
	"os/signal"
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
	"github.com/tutumagi/pitaya/engine/bc/baseapp"
	"github.com/tutumagi/pitaya/engine/common"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
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

	BootEntityType string
}

// Started -> ...
// Stop(system) -> stoppingMessage(user) ->  stateStopped(user)
// 										 ->  restarting(user)

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
		serverMode:       Cluster,
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

	app.server.Frontend = false
	app.server.Type = serverType
	app.serverMode = serverMode
	app.server.Metadata = serverMetadata
	app.messageEncoder = message.NewMessagesEncoder(app.config.GetBool("pitaya.handler.messages.compression"))
	configureMetrics(serverType)

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

	if err := common.RegisterModuleBefore(app.rpcServer, "rpcServer"); err != nil {
		logger.Log.Fatal("failed to register rpc server module: %s", err.Error())
	}
	if err := common.RegisterModuleBefore(app.rpcClient, "rpcClient"); err != nil {
		logger.Log.Fatal("failed to register rpc client module: %s", err.Error())
	}
	// set the service discovery as the last module to be started to ensure
	// all modules have been properly initialized before the server starts
	// receiving requests from other pitaya servers
	if err := common.RegisterModuleAfter(app.serviceDiscovery, "serviceDiscovery"); err != nil {
		logger.Log.Fatal("failed to register service discovery module: %s", err.Error())
	}

	app.router.SetServiceDiscovery(app.serviceDiscovery)

	// }

	actorSystem := actor.NewActorSystem()

	handlerService = NewAppProcessor(
		app.dieChan,
		app.serializer,
		app.server,
		app.messageEncoder,
		app.metricsReporters,
		app.rpcClient,
		app.rpcServer,
		app.serviceDiscovery,
		app.router,
		actorSystem,
	)

	app.rpcServer.SetPitayaServer(handlerService.remote)

	initSysRemotes()

	baseapp.Initialize(
		app.dieChan,
		app.rpcClient,
		app.serializer,
		app.serviceDiscovery,
		actorSystem,
		handlerService.remote,
	)

	periodicMetrics()

	// 开始连接层的事件循环
	listen()

	defer func() {
		timer.GlobalTicker.Stop()
		app.running = false
	}()

	sg := make(chan os.Signal, 1)
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
	common.ShutdownModules()
	// shutdownComponents()
}

func listen() {
	// startupComponents()
	// create global ticker instance, timer precision could be customized
	// by SetTimerPrecision
	timer.GlobalTicker = time.NewTicker(timer.Precision)

	logger.Log.Infof("starting server %s:%s", app.server.Type, app.server.ID)
	for i := 0; i < app.config.GetInt("pitaya.concurrency.handler.dispatch"); i++ {
		go handlerService.Dispatch(i)
	}

	common.StartModules()

	logger.Log.Info("all modules started!")

	app.running = true
}
