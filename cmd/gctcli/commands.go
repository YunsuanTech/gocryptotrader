package main

import (
	"fmt"
	"gocryptotrader/gctrpc"

	"github.com/urfave/cli/v2"
)

var startTime, endTime, orderingDirection string
var limit int

var monitorPriceCommand = &cli.Command{
	Name:   "monitorprice",
	Usage:  "监控指定交易对的价格",
	Action: monitorPrice,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "symbol",
			Usage: "交易对，例如 btcusdt",
		},
		&cli.Int64Flag{
			Name:  "timeout",
			Usage: "监控持续时间（秒），如果为 0 则一直监控直到出错",
		},
	},
}

var tradeTokenBySignalCommand = &cli.Command{
	Name:   "tradetokenbysignal",
	Usage:  "trade token by signal",
	Action: tradeTokenBySignal,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "token_address",
			Usage: "the address of the token to trade",
		},
		&cli.Float64Flag{
			Name:  "buy_price",
			Usage: "the price to buy the token",
		},
	},
}

var getAccountsCommand = &cli.Command{
	Name:   "getaccounts",
	Usage:  "gets GoCryptoTrader accounts",
	Action: getAccounts,
}

var getProfitLossCommand = &cli.Command{
	Name:   "getProfitLoss",
	Usage:  "gets GoCryptoTrader ProfitLoss",
	Action: getProfitLoss,
}

var buySOLTokenCommand = &cli.Command{
	Name:   "buySOLToken",
	Usage:  "gets GoCryptoTrader BuySOLToken",
	Action: BuySOLToken,
}

var stopSOLTokenMonitorCommand = &cli.Command{
	Name:   "stopSOLTokenMonitor",
	Usage:  "stops SOL token monitoring service",
	Action: StopSOLTokenMonitor,
}

var stopSignalMonitorCommand = &cli.Command{
	Name:   "stopsignalmonitor",
	Usage:  "stops signal monitoring service",
	Action: stopSignalMonitor,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "token_address",
			Usage: "the address of the token to trade",
		},
	},
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

func getProfitLoss(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.GetProfitLoss(c.Context,
		&gctrpc.GetProfitLossRequest{},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func BuySOLToken(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.BuySOLToken(c.Context,
		&gctrpc.BuySOLTokenRequest{},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func StopSOLTokenMonitor(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.StopSOLTokenMonitor(c.Context,
		&gctrpc.StopAutoTradeRequest{},
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

func stopSignalMonitor(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)
	tokenAddress := c.String("token_address")
	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.StopSignalMonitor(c.Context,
		&gctrpc.StopSignalMonitorRequest{
			TokenAddress: tokenAddress,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func tradeTokenBySignal(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	tokenAddress := c.String("token_address")
	buyPrice := c.Float64("buy_price")
	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.TradeTokenBySignal(c.Context,
		&gctrpc.TradeTokenBySignalRequest{
			TokenAddress: tokenAddress,
			BuyPrice:     buyPrice,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func monitorPrice(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	symbol := c.String("symbol")
	timeoutSeconds := c.Int64("timeout")

	if symbol == "" {
		return fmt.Errorf("交易对不能为空")
	}

	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.MonitorPrice(c.Context,
		&gctrpc.MonitorPriceRequest{
			Symbol:         symbol,
			TimeoutSeconds: timeoutSeconds,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}
