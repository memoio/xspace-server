package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/memoio/xspace-server/server"
	"github.com/urfave/cli/v2"
)

var XspaceServerCmd = &cli.Command{
	Name:  "server",
	Usage: "xspace server",
	Subcommands: []*cli.Command{
		xspaceServerRunCmd,
	},
}

var xspaceServerRunCmd = &cli.Command{
	Name:  "run",
	Usage: "run xspace server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"e"},
			Usage:   "input your port",
			Value:   "7890",
		},
		// &cli.StringFlag{
		// 	Name:  "sk",
		// 	Usage: "input your private key",
		// 	Value: "",
		// },
		&cli.StringFlag{
			Name:  "chain",
			Usage: "input chain name, e.g.(dev)",
			Value: "dev",
		},
		// &cli.StringFlag{
		// 	Name:  "ip",
		// 	Usage: "input meeda store node's ip address",
		// 	Value: "http://183.240.197.189:38082",
		// },
	},
	Action: func(ctx *cli.Context) error {
		port := ctx.String("port")
		// sk := ctx.String("sk")
		chain := ctx.String("chain")
		// ip := ctx.String("ip")

		// privateKey, err := crypto.HexToECDSA(sk)
		// if err != nil {
		// 	privateKey, err = crypto.GenerateKey()
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := server.NewServer(chain, port)
		if err != nil {
			log.Fatalf("new store node server: %s\n", err)
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		if err := srv.Shutdown(cctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}

		log.Println("Server exiting")

		return nil
	},
}
