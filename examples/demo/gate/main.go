package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/serialize/json"

	"github.com/tutumagi/pitaya/engine/components/gate"
)

func main() {
	defer gate.Shutdown()

	s := json.NewSerializer()
	conf := configApp()

	gate.SetSerializer(s)

	log.SetFlags(log.LstdFlags | log.Llongfile)

	// startWeb()

	t := acceptor.NewWSAcceptor(":3250")
	gate.AddAcceptor(t)

	gate.Configure("gate", map[string]string{}, conf)
	gate.Start()
}

func configApp() *viper.Viper {
	conf := viper.New()
	conf.SetEnvPrefix("chat") // allows using env vars in the CHAT_PITAYA_ format
	conf.SetDefault("gate.buffer.handler.localprocess", 15)
	conf.Set("gate.heartbeat.interval", "15s")
	conf.Set("gate.buffer.agent.messages", 32)
	conf.Set("gate.handler.messages.compression", false)
	return conf
}

// func startWeb() {
// 	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

// 	go http.ListenAndServe(":3251", nil)
// }
