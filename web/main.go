package main

import (
	"fmt"
	"os"

	"github.com/BlackCodes/micro-swagger/web/handler"
	"github.com/BlackCodes/micro-swagger/web/router"
	"github.com/gin-gonic/gin"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/web"

	_ "github.com/micro/go-plugins/transport/tcp/v2"
)

func main() {
	r := gin.Default()
	router.Route(r)
	var service = web.NewService(
		web.Name("go.micro.web.spider"),
		web.Version("0.0.1"),
		web.Flags(

			&cli.StringFlag{
				Name:    "basePath",
				Usage:   "basePath",
				EnvVars: []string{"BASEPATH"},
			},
		),
		web.Action(func(ctx *cli.Context) {
			if handler.Storage = ctx.String("basePath"); len(handler.Storage) == 0 {
				handler.Storage = "./mp-api/json"
			}
		}),
		web.Address(":9099"),
		web.Handler(r),
	)

	if err := service.Init(); err != nil {
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
