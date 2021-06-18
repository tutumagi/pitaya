package app

import (
	"fmt"

	"github.com/tutumagi/pitaya/cluster"
)

func AddServerOberver(listener cluster.SDListener) {
	sd, err := cluster.NewEtcdServiceDiscovery(
		GetConfig(),
		GetServer(),
		GetDieChan(),
	)
	if err != nil {
		panic("ctor etcd service err")
	}
	sd.AddListener(listener)
	SetServiceDiscoveryClient(sd)
}

func GetAnyServerByType(typ string) (*cluster.Server, error) {
	ss, err := GetServersByType(typ)
	if err != nil {
		return nil, err
	}
	for _, v := range ss {
		return v, nil
	}
	return nil, fmt.Errorf("找不到该服务 %s 当前所有服务:%v", typ, GetServers())
}
