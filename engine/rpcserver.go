package engine

import (
	context "context"
	errors "errors"
	"fmt"

	"gocryptotrader/common/crypto"
	"gocryptotrader/common/forward"

	"gocryptotrader/exchanges/request"

	"gocryptotrader/exchanges/token"
	"gocryptotrader/log"

	net "net"
	http "net/http"
	filepath "path/filepath"
	strings "strings"
	sync "sync"
	time "time"

	"google.golang.org/grpc/metadata"

	"gocryptotrader/exchanges/account"
	"gocryptotrader/gctrpc"
	"gocryptotrader/gctrpc/auth"
	"gocryptotrader/utils"

	transactionrecord "gocryptotrader/database/repository/transaction_record"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	errExchangeNotLoaded       = errors.New("exchange is not loaded/doesn't exist")
	errExchangeNotEnabled      = errors.New("exchange is not enabled")
	errExchangeBaseNotFound    = errors.New("cannot get exchange base")
	errInvalidArguments        = errors.New("invalid arguments received")
	errExchangeNameUnset       = errors.New("exchange name unset")
	errCurrencyPairUnset       = errors.New("currency pair unset")
	errInvalidTimes            = errors.New("invalid start and end times")
	errAssetTypeUnset          = errors.New("asset type unset")
	errDispatchSystem          = errors.New("dispatch system offline")
	errCurrencyNotEnabled      = errors.New("currency not enabled")
	errCurrencyNotSpecified    = errors.New("a currency must be specified")
	errCurrencyPairInvalid     = errors.New("currency provided is not found in the available pairs list")
	errNoTrades                = errors.New("no trades returned from supplied params")
	errNilRequestData          = errors.New("nil request data received, cannot continue")
	errNoAccountInformation    = errors.New("account information does not exist")
	errShutdownNotAllowed      = errors.New("shutting down this bot instance is not allowed via gRPC, please enable by command line flag --grpcshutdown or config.json field grpcAllowBotShutdown")
	errGRPCShutdownSignalIsNil = errors.New("cannot shutdown, gRPC shutdown channel is nil")
	errInvalidStrategy         = errors.New("invalid strategy")
	errSpecificPairNotEnabled  = errors.New("specified pair is not enabled")
)

// RPCServer struct
type RPCServer struct {
	gctrpc.UnimplementedGoCryptoTraderServiceServer
	*Engine
	// token监控相关字段
	tokenMonitorTimer *time.Ticker
	tokenMonitorDone  chan bool
	tokenMonitoring   bool

	// 信号交易监控相关字段
	signalMonitorTimer *time.Ticker
	signalMonitorDone  chan bool
	signalMonitoring   bool
	signalRequests     map[string]*gctrpc.TradeTokenBySignalRequest // 使用map存储多个代币地址的监控请求
	signalRequestMutex sync.RWMutex                                 // 用于保护signalRequests的互斥锁
}

func (s *RPCServer) authenticateClient(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, errors.New("unable to extract metadata")
	}

	authStr, ok := md["authorization"]
	if !ok {
		return ctx, errors.New("authorization header missing")
	}

	if !strings.Contains(authStr[0], "Basic") {
		return ctx, errors.New("basic not found in authorization header")
	}

	decoded, err := crypto.Base64Decode(strings.Split(authStr[0], " ")[1])
	if err != nil {
		return ctx, errors.New("unable to base64 decode authorization header")
	}

	cred := strings.Split(string(decoded), ":")
	username := cred[0]
	password := cred[1]

	if username != s.Config.RemoteControl.Username ||
		password != s.Config.RemoteControl.Password {
		return ctx, errors.New("username/password mismatch")
	}

	if _, ok := md["verbose"]; ok {
		ctx = request.WithVerbose(ctx)
	}
	return ctx, nil
}

// StartRPCServer starts a gRPC server with TLS auth
func StartRPCServer(engine *Engine) {
	targetDir := utils.GetTLSDir(engine.Settings.DataDir)
	if err := CheckCerts(targetDir); err != nil {
		log.Errorf(log.GRPCSys, "gRPC CheckCerts failed. err: %s\n", err)
		return
	}

	log.Debugf(log.GRPCSys, "gRPC server support enabled. Starting gRPC server on https://%v.\n", engine.Config.RemoteControl.GRPC.ListenAddress)

	lis, err := net.Listen("tcp", engine.Config.RemoteControl.GRPC.ListenAddress)
	if err != nil {
		log.Errorf(log.GRPCSys, "gRPC server failed to bind to port: %s", err)
		return
	}

	creds, err := credentials.NewServerTLSFromFile(filepath.Join(targetDir, "cert.pem"), filepath.Join(targetDir, "key.pem"))

	if err != nil {
		log.Errorf(log.GRPCSys, "gRPC server could not load TLS keys: %s\n", err)
		return
	}

	s := RPCServer{Engine: engine}
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(grpcauth.UnaryServerInterceptor(s.authenticateClient)),
		grpc.StreamInterceptor(grpcauth.StreamServerInterceptor(s.authenticateClient)),
	}
	server := grpc.NewServer(opts...)
	gctrpc.RegisterGoCryptoTraderServiceServer(server, &s)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Errorf(log.GRPCSys, "gRPC server failed to serve: %s\n", err)
			return
		}
	}()

	log.Debugln(log.GRPCSys, "gRPC server started!")

	if s.Settings.EnableGRPCProxy {
		s.StartRPCRESTProxy()
	}
}

// StartRPCRESTProxy starts a gRPC proxy
func (s *RPCServer) StartRPCRESTProxy() {
	log.Debugf(log.GRPCSys, "gRPC proxy server support enabled. Starting gRPC proxy server on https://%v.\n", s.Config.RemoteControl.GRPC.GRPCProxyListenAddress)

	targetDir := utils.GetTLSDir(s.Settings.DataDir)
	certFile := filepath.Join(targetDir, "cert.pem")
	keyFile := filepath.Join(targetDir, "key.pem")
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Errorf(log.GRPCSys, "Unable to start gRPC proxy. Err: %s\n", err)
		return
	}

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(auth.BasicAuth{
			Username: s.Config.RemoteControl.Username,
			Password: s.Config.RemoteControl.Password,
		}),
	}
	err = gctrpc.RegisterGoCryptoTraderServiceHandlerFromEndpoint(context.Background(),
		mux, s.Config.RemoteControl.GRPC.ListenAddress, opts)
	if err != nil {
		log.Errorf(log.GRPCSys, "Failed to register gRPC proxy. Err: %s\n", err)
		return
	}

	go func() {
		server := &http.Server{
			Addr:              s.Config.RemoteControl.GRPC.GRPCProxyListenAddress,
			ReadHeaderTimeout: time.Minute,
			ReadTimeout:       time.Minute,
			Handler:           s.authClient(mux),
		}

		if err = server.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Errorf(log.GRPCSys, "gRPC proxy server failed to serve: %s\n", err)
			return
		}
	}()

	log.Debugln(log.GRPCSys, "gRPC proxy server started!")
}

func (s *RPCServer) authClient(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != s.Config.RemoteControl.Username || password != s.Config.RemoteControl.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Access denied", http.StatusUnauthorized)
			log.Warnf(log.GRPCSys, "gRPC proxy server unauthorised access attempt. IP: %s Path: %s\n", r.RemoteAddr, r.URL.Path)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// GetAccounts 获取所有账户信息
func (s *RPCServer) GetAccounts(ctx context.Context, req *gctrpc.GetAccountsRequest) (*gctrpc.GetAccountsResponse, error) {

	accountManager := account.New(s.Config)

	accounts, err := accountManager.Accounts()
	if err != nil {
		return nil, err
	}

	response := &gctrpc.GetAccountsResponse{}
	for _, acc := range accounts {
		response.Accounts = append(response.Accounts, &gctrpc.Account{
			Name:              acc.Name,
			Address:           acc.Address,
			ExchangeAddressId: acc.ExchangeAddressID,
			ZkAddressId:       acc.ZkAddressID,
			F4AddressId:       acc.F4AddressID,
			OtAddressId:       acc.OTAddressID,
			Cipher:            acc.Cipher,
			Layer:             int32(acc.Layer),
			Owner:             acc.Owner,
			ChainName:         acc.ChainName,
		})
	}

	return response, nil
}

// GetTokenPrice 获取代币价格信息
func (s *RPCServer) GetTokenPrice(ctx context.Context, req *gctrpc.GetTokenPriceRequest) (*gctrpc.GetTokenPriceResponse, error) {
	if req.TokenAddress == "" {
		return nil, errors.New("token address cannot be empty")
	}

	tokenPrice, err := token.GetTokenPrice(req.TokenAddress)

	if err != nil {
		return nil, err
	}

	response := &gctrpc.GetTokenPriceResponse{
		TokenPrice: &gctrpc.TokenPrice{
			Address:  tokenPrice.Address,
			UsdPrice: tokenPrice.USDPrice,
			SolPrice: tokenPrice.SOLPrice,
			LastUpdate: &gctrpc.Timestamp{
				Seconds: tokenPrice.LastUpdate.Unix(),
				Nanos:   int32(tokenPrice.LastUpdate.Nanosecond()),
			},
		},
	}

	return response, nil
}

// Crypto 实现加密服务
func (s *RPCServer) Crypto(ctx context.Context, req *gctrpc.CryptoRequest) (*gctrpc.CryptoResponse, error) {
	if req.Plaintext == "" {
		return nil, errors.New("enter a partial private key")
	}

	accountManager := account.New(s.Config)

	ciphertext, err := accountManager.Crypto(req.Plaintext)
	if err != nil {
		return nil, err
	}

	return &gctrpc.CryptoResponse{
		Ciphertext: ciphertext,
	}, nil
}

// TransferSOL 实现SOL代币批量转发服务
func (s *RPCServer) TransferSOL(ctx context.Context, req *gctrpc.TransferSOLRequest) (*gctrpc.TransferSOLResponse, error) {
	if req == nil {
		return nil, errNilRequestData
	}

	if req.Address == "" {
		return nil, errors.New("address cannot be empty")
	}

	// 获取账户管理器
	accountManager := account.New(s.Config)

	// 获取私钥
	privateKey, err := accountManager.PrivateKey(req.Address)
	if err != nil {
		return nil, fmt.Errorf("获取私钥失败: %w", err)
	}

	// 获取转发管理器
	forwardManager := forward.New(s.Config)

	// 从文件读取目标地址列表
	addresses, err := forward.ReadAddressesFromFile(s.Config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("读取地址列表失败: %w", err)
	}

	// 创建转发请求
	forwardReq := &forward.ForwardRequest{
		PrivateKeyStr: privateKey,
		Addresses:     addresses,
		Config:        forward.DefaultConfig(),
		AmountSOL:     1,
	}

	// 执行转发
	txSignatures, err := forwardManager.TransferSOL(ctx, forwardReq)
	if err != nil {
		return nil, err
	}

	return &gctrpc.TransferSOLResponse{
		TxSignatures: txSignatures,
	}, nil
}

// TransferToken 实现代币批量转发服务
func (s *RPCServer) TransferToken(ctx context.Context, req *gctrpc.TransferTokenRequest) (*gctrpc.TransferTokenResponse, error) {
	if req == nil {
		return nil, errNilRequestData
	}

	if req.Address == "" {
		return nil, errors.New("address cannot be empty")
	}

	if req.TokenMint == "" {
		return nil, errors.New("token mint cannot be empty")
	}

	// 获取账户管理器
	accountManager := account.New(s.Config)

	// 获取私钥
	privateKey, err := accountManager.PrivateKey(req.Address)
	if err != nil {
		return nil, fmt.Errorf("获取私钥失败: %w", err)
	}

	// 获取转发管理器
	forwardManager := forward.New(s.Config)

	// 从文件读取目标地址列表
	addresses, err := forward.ReadAddressesFromFile(s.Config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("读取地址列表失败: %w", err)
	}

	// 创建转发请求
	forwardReq := &forward.TokenForwardRequest{
		PrivateKeyStr: privateKey,
		TokenMint:     req.TokenMint,
		Addresses:     addresses,
		IsToken2022:   false,
		Config:        forward.DefaultConfig(),
		Amount:        100,
	}

	// 执行转发
	txSignatures, err := forwardManager.TransferToken(ctx, forwardReq)
	if err != nil {
		return nil, err
	}

	return &gctrpc.TransferTokenResponse{
		TxSignatures: txSignatures,
	}, nil
}

// BuySOLToken 启动SOL代币监控服务
func (s *RPCServer) BuySOLToken(ctx context.Context, req *gctrpc.BuySOLTokenRequest) (*gctrpc.BuySOLTokenResponse, error) {
	// 如果已经在监控中，返回提示信息
	if s.tokenMonitoring {
		return &gctrpc.BuySOLTokenResponse{
			Success: false,
			Message: "SOL代币监控服务已经在运行中",
		}, nil
	}

	// 初始化监控通道
	s.tokenMonitorDone = make(chan bool)
	// 创建定时器，每秒执行一次
	s.tokenMonitorTimer = time.NewTicker(10 * time.Minute)

	// 标记为监控中
	s.tokenMonitoring = true
	fetcher, err := token.NewGMGNFetcher()
	if err != nil {
		log.Errorf(log.GRPCSys, "创建GMGNFetcher实例失败: %v", err)

	}

	// 启动监控协程
	go func() {
		for {
			select {
			case <-s.tokenMonitorTimer.C:

				// 使用WaitGroup来等待两个操作都完成
				var wg sync.WaitGroup
				wg.Add(2)

				// 并发执行买入操作
				go func() {
					defer wg.Done()
					// 创建GMGNFetcher实例

					err := token.BuySOLToken(fetcher)
					if err != nil {
						log.Errorf(log.GRPCSys, "执行SOL代币买入操作失败: %v", err)
					}
				}()

				// 并发执行卖出操作
				go func() {
					defer wg.Done()
					err := token.SellSOLToken(fetcher)
					if err != nil {
						log.Errorf(log.GRPCSys, "执行SOL代币卖出操作失败: %v", err)
					}
				}()

				// 等待两个操作都完成
				wg.Wait()

			case <-s.tokenMonitorDone:
				// 收到停止信号，退出协程
				return
			}
		}
	}()

	return &gctrpc.BuySOLTokenResponse{
		Success: true,
		Message: "SOL代币监控服务已成功启动",
	}, nil
}

// StopSOLTokenMonitor 停止SOL代币监控服务
func (s *RPCServer) StopSOLTokenMonitor(ctx context.Context, req *gctrpc.StopAutoTradeRequest) (*gctrpc.StopAutoTradeResponse, error) {
	// 如果没有在监控中，返回提示信息
	if !s.tokenMonitoring {
		return &gctrpc.StopAutoTradeResponse{
			Success: false,
			Message: "SOL代币监控服务未启动",
		}, nil
	}

	// 停止定时器
	s.tokenMonitorTimer.Stop()
	// 发送停止信号
	s.tokenMonitorDone <- true
	// 关闭通道
	close(s.tokenMonitorDone)

	// 标记为未监控
	s.tokenMonitoring = false

	return &gctrpc.StopAutoTradeResponse{
		Success: true,
		Message: "SOL代币监控服务已成功停止",
	}, nil
}

// GetProfitLoss 获取所有代币的盈亏情况
func (s *RPCServer) GetProfitLoss(ctx context.Context, req *gctrpc.GetProfitLossRequest) (*gctrpc.GetProfitLossResponse, error) {
	// 调用 transaction_record 包的 GetProfitLoss 函数获取盈亏数据
	profitLosses, err := transactionrecord.GetProfitLoss(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为 gRPC 响应格式
	response := &gctrpc.GetProfitLossResponse{}
	for _, pl := range profitLosses {
		response.TokenProfitLosses = append(response.TokenProfitLosses, &gctrpc.TokenProfitLoss{
			TokenAddress:    pl.TokenAddress,
			TotalBuyAmount:  pl.TotalBuyAmount,
			TotalSellAmount: pl.TotalSellAmount,
			ProfitLoss:      pl.ProfitLoss,
		})
	}

	return response, nil
}

// TradeTokenBySignal 根据信号交易代币
func (s *RPCServer) TradeTokenBySignal(ctx context.Context, req *gctrpc.TradeTokenBySignalRequest) (*gctrpc.TradeTokenBySignalResponse, error) {
	if req == nil {
		return nil, errNilRequestData
	}

	if req.TokenAddress == "" {
		return nil, errors.New("token address cannot be empty")
	}

	if req.BuyPrice <= 0 {
		return nil, errors.New("buy price must be greater than zero")
	}

	// 检查是否已经在监控这个代币地址
	s.signalRequestMutex.RLock()
	if s.signalRequests != nil {
		if _, exists := s.signalRequests[req.TokenAddress]; exists {
			s.signalRequestMutex.RUnlock()
			return &gctrpc.TradeTokenBySignalResponse{
				Success: false,
				Message: fmt.Sprintf("代币地址 %s 已经在监控中", req.TokenAddress),
			}, nil
		}
	}
	s.signalRequestMutex.RUnlock()

	// 如果监控服务尚未启动，则初始化
	if !s.signalMonitoring {
		// 初始化监控通道和请求映射
		s.signalMonitorDone = make(chan bool)
		s.signalRequestMutex.Lock()
		s.signalRequests = make(map[string]*gctrpc.TradeTokenBySignalRequest)
		s.signalRequestMutex.Unlock()
		// 创建定时器，每5分钟执行一次
		s.signalMonitorTimer = time.NewTicker(60 * time.Minute)

		// 标记为监控中
		s.signalMonitoring = true

		// 启动监控协程
		go func() {
			for {
				select {
				case <-s.signalMonitorTimer.C:
					// 获取当前所有监控的代币请求
					s.signalRequestMutex.RLock()
					requests := make(map[string]*gctrpc.TradeTokenBySignalRequest, len(s.signalRequests))
					for addr, req := range s.signalRequests {
						requests[addr] = req
					}
					s.signalRequestMutex.RUnlock()

					// 遍历所有代币请求并执行交易
					for _, req := range requests {
						var err error
						if req.TradeRatio > 0 {
							// 如果提供了交易比例，则传递该参数
							err = token.TradeTokenBySignal(req.TokenAddress, req.BuyPrice, req.TradeRatio)
						} else {
							// 否则使用默认比例
							err = token.TradeTokenBySignal(req.TokenAddress, req.BuyPrice)
						}

						if err != nil {
							log.Errorf(log.GRPCSys, "执行信号交易失败，代币地址: %s, 错误: %v", req.TokenAddress, err)
						}
					}

				case <-s.signalMonitorDone:
					// 收到停止信号，退出协程
					return
				}
			}
		}()
	}

	// 添加新的代币监控请求
	s.signalRequestMutex.Lock()
	s.signalRequests[req.TokenAddress] = req
	s.signalRequestMutex.Unlock()

	return &gctrpc.TradeTokenBySignalResponse{
		Success: true,
		Message: fmt.Sprintf("信号交易监控服务已成功添加代币地址: %s，将每5分钟执行一次交易检查", req.TokenAddress),
	}, nil
}

// StopSignalMonitor 停止信号交易监控服务
func (s *RPCServer) StopSignalMonitor(ctx context.Context, req *gctrpc.StopSignalMonitorRequest) (*gctrpc.StopSignalMonitorResponse, error) {
	// 如果没有在监控中，返回提示信息
	if !s.signalMonitoring {
		return &gctrpc.StopSignalMonitorResponse{
			Success: false,
			Message: "信号交易监控服务未启动",
		}, nil
	}

	// 如果请求中指定了代币地址，则只停止该代币的监控
	if req.TokenAddress != "" {
		s.signalRequestMutex.Lock()
		if _, exists := s.signalRequests[req.TokenAddress]; exists {
			delete(s.signalRequests, req.TokenAddress)
			remaining := len(s.signalRequests)
			s.signalRequestMutex.Unlock()

			// 如果没有剩余监控的代币，则完全停止监控服务
			if remaining == 0 {
				// 停止定时器
				s.signalMonitorTimer.Stop()
				// 发送停止信号
				s.signalMonitorDone <- true
				// 关闭通道
				close(s.signalMonitorDone)
				// 标记为未监控
				s.signalMonitoring = false
			}

			return &gctrpc.StopSignalMonitorResponse{
				Success: true,
				Message: fmt.Sprintf("已停止监控代币地址: %s，剩余监控代币数量: %d", req.TokenAddress, remaining),
			}, nil
		}
		s.signalRequestMutex.Unlock()
		return &gctrpc.StopSignalMonitorResponse{
			Success: false,
			Message: fmt.Sprintf("代币地址 %s 未在监控列表中", req.TokenAddress),
		}, nil
	}

	// 如果未指定代币地址，则停止所有监控
	// 停止定时器
	s.signalMonitorTimer.Stop()
	// 发送停止信号
	s.signalMonitorDone <- true
	// 关闭通道
	close(s.signalMonitorDone)

	// 清理请求数据
	s.signalRequestMutex.Lock()
	s.signalRequests = nil
	s.signalRequestMutex.Unlock()
	// 标记为未监控
	s.signalMonitoring = false

	return &gctrpc.StopSignalMonitorResponse{
		Success: true,
		Message: "信号交易监控服务已成功停止所有代币监控",
	}, nil
}

// // MonitorPrice 实现价格监控服务
// func (s *RPCServer) MonitorPrice(ctx context.Context, req *gctrpc.MonitorPriceRequest) (*gctrpc.MonitorPriceResponse, error) {
// 	if req == nil {
// 		return nil, errNilRequestData
// 	}

// 	if req.Symbol == "" {
// 		return nil, errors.New("交易对不能为空")
// 	}

// 	// 设置超时时间
// 	var timeout time.Duration
// 	if req.TimeoutSeconds > 0 {
// 		timeout = time.Duration(req.TimeoutSeconds) * time.Second
// 	}

// 	// 创建一个自定义回调函数，用于处理行情数据并根据需要保存到数据库
// 	callback := func(ticker handling.BinanceTicker) {
// 		// 使用默认处理函数显示行情数据
// 		handling.DefaultTickerHandler(ticker)
// 	}

// 	// 启动一个新的 goroutine 来执行监控，这样不会阻塞 RPC 调用
// 	go func() {
// 		err := handling.MonitorPrice(req.Symbol, callback, timeout)
// 		if err != nil {
// 			log.Errorf(log.GRPCSys, "价格监控服务出错: %v", err)
// 		}
// 	}()

// 	return &gctrpc.MonitorPriceResponse{
// 		Success: true,
// 		Message: fmt.Sprintf("已启动对 %s 的价格监控服务", req.Symbol),
// 	}, nil
// }
