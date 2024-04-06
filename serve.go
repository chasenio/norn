package main

import (
	"fmt"
	"github.com/kentio/norn/internal/service"
	"github.com/kentio/norn/web"
	"github.com/sirupsen/logrus"
)

func main() {
	conf, err := service.NewConfig()
	if err != nil {
		panic(err)
	}
	conf.Output()
	app := web.NewApp(conf)
	logrus.Infof("server is running on port %s", conf.HTTPPort)
	err = app.Run(fmt.Sprintf(":%s", conf.HTTPPort))
	if err != nil {
		panic(err)
	}
}
