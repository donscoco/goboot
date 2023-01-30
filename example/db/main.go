package main

import (
	"fmt"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/core"
	"log"
	"time"
)

func main() {
	config.ConfigFilePath = "/data/github/gomodule-demo/goboot/config-demo.json"

	fmt.Println(config.ConfigFilePath)

	core.GoCore = core.NewCore()
	core.GoCore.OnStart(
		core.InitDBManager,
		TestService,
	)
	core.GoCore.OnStop()
	core.GoCore.Boot()

}

func TestService() (err error) {

	//p := core.GoCore.DB.MySQL["ironhead-mysql"]
	mp, err := core.GoCore.DB.GetMySQL("ironhead-mysql")
	if err != nil {
		log.Fatal(err)
	}
	row, err := mp.Session.Query("SELECT * FROM test.iron_student")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(row.Columns())

	rp, err := core.GoCore.DB.GetRedis("ironhead-redis")
	if err != nil {
		log.Fatal(err)
	}
	cmd := rp.SigleClient.SetNX("testkey", "testval", 60*time.Second)
	log.Println(cmd.Result())

	return nil
}
