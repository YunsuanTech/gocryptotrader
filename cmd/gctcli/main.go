package main

import (
	context "context"
	encoding_json "encoding/json"
	fmt "fmt"
	log "log"
	os "os"
	filepath "path/filepath"
	runtime "runtime"
	time "time"

	"gocryptotrader/common"
	"gocryptotrader/core"
	"gocryptotrader/gctrpc/auth"
	"gocryptotrader/signaler"

	"github.com/urfave/cli/v2"
	grpc "google.golang.org/grpc"
	credentials "google.golang.org/grpc/credentials"
	metadata "google.golang.org/grpc/metadata"
)

var (
	host          string
	username      string
	password      string
	pairDelimiter string
	certPath      string
	timeout       time.Duration
	verbose       bool
	ignoreTimeout bool
)

const defaultTimeout = time.Second * 30

func jsonOutput(in interface{}) {
	j, err := encoding_json.MarshalIndent(in, "", " ")
	if err != nil {
		return
	}
	fmt.Print(string(j))
}

func setupClient(c *cli.Context) (*grpc.ClientConn, context.CancelFunc, error) {
	creds, err := credentials.NewClientTLSFromFile(certPath, "")
	if err != nil {
		return nil, nil, err
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(auth.BasicAuth{
			Username: username,
			Password: password,
		}),
	}

	var cancel context.CancelFunc
	if !ignoreTimeout {
		c.Context, cancel = context.WithTimeout(c.Context, timeout)
	}

	if verbose {
		c.Context = metadata.AppendToOutgoingContext(c.Context, "verbose", "true")
	}
	conn, err := grpc.Dial(host, opts...)
	return conn, cancel, err
}

func main() {
	app := cli.NewApp()
	app.Name = "gctcli"
	app.Version = core.Version(true)
	app.EnableBashCompletion = true
	app.Usage = "command line interface for managing the gocryptotrader daemon"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "rpchost",
			Value:       "localhost:9052",
			Usage:       "the gRPC host to connect to",
			Destination: &host,
		},
		&cli.StringFlag{
			Name:        "rpcuser",
			Value:       "admin",
			Usage:       "the gRPC username",
			Destination: &username,
		},
		&cli.StringFlag{
			Name:        "rpcpassword",
			Value:       "Password",
			Usage:       "the gRPC password",
			Destination: &password,
		},
		&cli.StringFlag{
			Name:        "delimiter",
			Value:       "-",
			Usage:       "the default currency pair delimiter used to standardise currency pair input",
			Destination: &pairDelimiter,
		},
		&cli.StringFlag{
			Name:        "cert",
			Value:       filepath.Join(common.GetDefaultDataDir(runtime.GOOS), "tls", "cert.pem"),
			Usage:       "the path to TLS cert of the gRPC server",
			Destination: &certPath,
		},
		&cli.DurationFlag{
			Name:        "timeout",
			Value:       defaultTimeout,
			Usage:       "the default context timeout value for requests",
			Destination: &timeout,
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "allows the request to generate a more verbose outputs server side",
			Destination: &verbose,
		},
		&cli.BoolFlag{
			Name:        "ignoretimeout",
			Aliases:     []string{"it"},
			Usage:       "ignores the context timeout for requests",
			Destination: &ignoreTimeout,
		},
	}
	app.Commands = []*cli.Command{
		getAccountsCommand,
		getTokenPriceCommand,
		cryptoCommand,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Capture cancel for interrupt
		signaler.WaitForInterrupt()
		cancel()
		fmt.Println("rpc process interrupted")
		os.Exit(1)
	}()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
