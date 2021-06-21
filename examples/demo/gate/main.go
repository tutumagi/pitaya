package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/serialize/json"

	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/components/gate"
)

func main() {
	defer gate.Shutdown()

	s := json.NewSerializer()
	conf := configApp()

	gate.SetSerializer(s)

	log.SetFlags(log.LstdFlags | log.Llongfile)

	t := acceptor.NewWSAcceptor(":3250")
	gate.AddAcceptor(t)

	gate.Configure(metapart.GateAppSvr, map[string]string{}, conf)
	gate.Start()
}

func configApp() *viper.Viper {
	conf := viper.New()
	// conf.SetEnvPrefix("chat") // allows using env vars in the CHAT_PITAYA_ format
	conf.SetDefault("pitaya.buffer.handler.localprocess", 15)
	conf.Set("pitaya.heartbeat.interval", "15s")
	conf.Set("pitaya.buffer.agent.messages", 32)
	conf.Set("pitaya.handler.messages.compression", false)

	conf.Set("pitaya.bootentity", "account")
	return conf
}
