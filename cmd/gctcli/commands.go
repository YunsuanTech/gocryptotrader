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

var getTokenPriceCommand = &cli.Command{
	Name:   "gettokenprice",
	Usage:  "gets token price information",
	Action: getTokenPrice,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "token_address",
			Usage: "the address of the token to get price for",
		},
	},
}

var cryptoCommand = &cli.Command{
	Name:   "crypto",
	Usage:  "encrypts the provided plaintext",
	Action: crypto,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "plaintext",
			Usage: "the text to be encrypted",
		},
	},
}

var transferSOLCommand = &cli.Command{
	Name:   "transfersol",
	Usage:  "transfer SOL tokens to multiple addresses",
	Action: transferSOL,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Usage: "the source address to transfer SOL from",
		},
	},
}

var transferTokenCommand = &cli.Command{
	Name:   "transfertoken",
	Usage:  "transfer tokens to multiple addresses",
	Action: transferToken,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Usage: "the source address to transfer tokens from",
		},
		&cli.StringFlag{
			Name:  "token_mint",
			Usage: "the token mint address",
		},
	},
}

func transferToken(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	address := c.String("address")
	tokenMint := c.String("token_mint")
	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.TransferToken(c.Context,
		&gctrpc.TransferTokenRequest{
			Address:   address,
			TokenMint: tokenMint,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func transferSOL(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	address := c.String("address")
	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.TransferSOL(c.Context,
		&gctrpc.TransferSOLRequest{
			Address: address,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
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

func getTokenPrice(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	tokenAddress := c.String("token_address")
	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.GetTokenPrice(c.Context,
		&gctrpc.GetTokenPriceRequest{
			TokenAddress: tokenAddress,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func crypto(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	plaintext := c.String("plaintext")
	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.Crypto(c.Context,
		&gctrpc.CryptoRequest{
			Plaintext: plaintext,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}
