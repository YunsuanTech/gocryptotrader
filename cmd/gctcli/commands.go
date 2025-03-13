package main

import (
	"gocryptotrader/gctrpc"

	"github.com/urfave/cli/v2"
)

var startTime, endTime, orderingDirection string
var limit int

var getAccountsCommand = &cli.Command{
	Name:   "getaccounts",
	Usage:  "gets GoCryptoTrader accounts",
	Action: getAccounts,
}

func getAccounts(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.GetAccounts(c.Context,
		&gctrpc.GetAccountsRequest{},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}
