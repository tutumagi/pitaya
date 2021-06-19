package main

import (
	"context"
	"log"
	"net/http"

	"strings"

	"github.com/spf13/viper"

	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/config"
	"github.com/tutumagi/pitaya/engine/bc"
	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"github.com/tutumagi/pitaya/engine/components/app"
	"github.com/tutumagi/pitaya/groups"
	"github.com/tutumagi/pitaya/serialize/json"
	"github.com/tutumagi/pitaya/timer"
)

type (
	// Room represents a component that contains a bundle of room related handler
	// like Join/Message
	Room struct {
		basepart.Entity
		timer *timer.Timer
	}

	// UserMessage represents a message that user sent
	UserMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// NewUser message will be received when new user join room
	NewUser struct {
		Content string `json:"content"`
	}

	// AllMembers contains all members uid
	AllMembers struct {
		Members []string `json:"members"`
	}

	// JoinResponse represents the result of joining room
	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

func main() {
	defer app.Shutdown()

	s := json.NewSerializer()
	conf := configApp()

	app.SetSerializer(s)
	gsi := groups.NewMemoryGroupService(config.NewConfig(conf))
	app.InitGroups(gsi)
	err := app.GroupCreate(context.Background(), "room")
	if err != nil {
		panic(err)
	}

	// rewrite component and handler name
	desc := bc.RegisterService("room", &RoomService{})
	desc.Routers.Register(&Room{},
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)

	log.SetFlags(log.LstdFlags | log.Llongfile)

	_ = basepart.CreateService("room")

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	go http.ListenAndServe(":3251", nil)

	t := acceptor.NewWSAcceptor(":3250")
	app.AddAcceptor(t)

	app.Configure(true, "chat", app.Cluster, map[string]string{}, conf)
	app.Start()
}

func configApp() *viper.Viper {
	conf := viper.New()
	conf.SetEnvPrefix("chat") // allows using env vars in the CHAT_PITAYA_ format
	conf.SetDefault("app.buffer.handler.localprocess", 15)
	conf.Set("app.heartbeat.interval", "15s")
	conf.Set("app.buffer.agent.messages", 32)
	conf.Set("app.handler.messages.compression", false)
	return conf
}
