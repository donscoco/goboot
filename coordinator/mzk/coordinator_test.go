package mzk

import (
	"fmt"
	"goboot/config"
	"goboot/log/mlog"
	"log"
	"testing"
)

func TestCreateCoordinator(t *testing.T) {
	config.ConfigFilePath = "../../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	mlog.InitLoggerByConfig(conf, "/core/log")

	zkc, err := CreateCoordinator(conf, "/core/coordinator/zookeeper")
	if err != nil {
		log.Fatal(err)
	}
	zkc.Start()
	n, err := zkc.CreateNode("/c1", []byte("test"), false)
	if err != nil {
		log.Fatal(err)
	}
	d, err := n.Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(d))

	n, err = n.Create("/d1", []byte("test2"), false)
	if err != nil {
		log.Fatal(err)
	}
	d, err = n.Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(d))
	zkc.Stop()
}
