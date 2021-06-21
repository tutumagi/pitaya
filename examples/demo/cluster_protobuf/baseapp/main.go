package main

import (
	"flag"

	"github.com/tutumagi/pitaya/engine/components/app"
	"github.com/tutumagi/pitaya/examples/demo/cluster_protobuf/baseapp/cmd/route"
	"github.com/tutumagi/pitaya/serialize/protobuf"
)

func configureBackend() {
	route.RegisterRoute()
}

func main() {
	svType := flag.String("type", "connector", "the server type")
	isFrontend := flag.Bool("frontend", true, "if server is frontend")

	flag.Parse()

	defer app.Shutdown()

	ser := protobuf.NewSerializer()

	app.SetSerializer(ser)

	configureBackend()

	app.Configure(*isFrontend, *svType, app.Cluster, map[string]string{})
	app.Start()
}
