package main

import (
	"fmt"
	"goboot/config"
	"goboot/core"
	"goboot/log/mlog"
	"goboot/mq/mkafka"
	"net/http"
	_ "net/http/pprof"
)

func main() {

	config.ConfigFilePath = "../../config-demo.json"

	core.GoCore = core.NewCore()
	core.GoCore.OnStart(work, ppfunc)
	core.GoCore.OnStop()
	core.GoCore.Boot()

	//sig := make(chan os.Signal)
	//signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//<-sig

}

func work() error {
	consumer, err := mkafka.CreateConsumer(core.GoCore.GetConf(), "/core/mq/kafka/consumer-config")
	if err != nil {
		mlog.Errorf("%s", err.Error())
	}
	fmt.Println(consumer)
	consumer.Start()

	go func() {
		for {
			select {
			case msg := <-consumer.Output():
				fmt.Println("recv:", msg.Value)
			}
		}
	}()
	return nil
}

func ppfunc() error {
	fmt.Print(http.ListenAndServe(":6060", nil))

	return nil
}
