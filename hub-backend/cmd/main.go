package main

import (
	"github.com/urfave/cli/v2"
	hub_backend "github.com/zyjblockchain/hub-backend"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := &cli.App{
		Name: "mirror-hub",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "mysql", Value: "root@tcp(127.0.0.1:3306)/mirror_hub?charset=utf8mb4&parseTime=True&loc=Local", Usage: "mysql dsn", EnvVars: []string{"MYSQL"}},
			&cli.StringFlag{Name: "ar_node", Value: "https://arweave.net", EnvVars: []string{"AR_NODE"}},
			&cli.StringFlag{Name: "sentry", Value: "", Usage: "code runtime environment", EnvVars: []string{"SENTRY"}},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// run iweave logic code
	h := hub_backend.New(c.String("mysql"), c.String("ar_node"))
	h.Run()

	<-signals
	h.Close()

	return nil
}
