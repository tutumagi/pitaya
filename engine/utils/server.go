package utils

import (
	"fmt"

	"github.com/tutumagi/pitaya"
	"github.com/tutumagi/pitaya/cluster"
)

func AddServerOberver(listener cluster.SDListener) {
	sd, err := cluster.NewEtcdServiceDiscovery(
		pitaya.GetConfig(),
		pitaya.GetServer(),
		pitaya.GetDieChan(),
	)
	if err != nil {
		panic("ctor etcd service err")
	}
	sd.AddListener(listener)
	pitaya.SetServiceDiscoveryClient(sd)
}

func GetAnyServerByType(typ string) (*cluster.Server, error) {
	ss, err := pitaya.GetServersByType(typ)
	if err != nil {
		return nil, err
	}
	for _, v := range ss {
		return v, nil
	}
	return nil, fmt.Errorf("找不到该服务 %s 当前所有服务:%v", typ, pitaya.GetServers())
}
