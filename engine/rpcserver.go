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
	s.tokenMonitorTimer = time.NewTicker(1 * time.Second)

	// 标记为监控中
	s.tokenMonitoring = true

	// 启动监控协程
	go func() {
		for {
			select {
			case <-s.tokenMonitorTimer.C:
				// 创建token客户端
				client := token.NewClient()

				// 使用WaitGroup来等待两个操作都完成
				var wg sync.WaitGroup
				wg.Add(2)

				// 并发执行买入操作
				go func() {
					defer wg.Done()
					err := token.BuySOLToken(client)
					if err != nil {
						log.Errorf(log.GRPCSys, "执行SOL代币买入操作失败: %v", err)
					}
				}()

				// 并发执行卖出操作
				go func() {
					defer wg.Done()
					err := token.SellSOLToken(client)
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

// // StartAutoTrade 启动自动交易服务
// func (s *RPCServer) StartAutoTrade(ctx context.Context, req *gctrpc.StartAutoTradeRequest) (*gctrpc.StartAutoTradeResponse, error) {
// 	if req == nil {
// 		return nil, errNilRequestData
// 	}

// 	if req.Address == "" {
// 		return nil, errors.New("address cannot be empty")
// 	}

// 	// 获取账户管理器
// 	accountManager := account.New(s.Config)

// 	// 获取私钥
// 	privateKey, err := accountManager.PrivateKey(req.Address)
// 	if err != nil {
// 		return nil, fmt.Errorf("获取私钥失败: %w", err)
// 	}

// 	// 创建自动交易配置
// 	autoConfig := token.DefaultAutoTradeConfig()
// 	autoConfig.PrivateKey = privateKey

// 	// 如果用户指定了检查间隔时间，则使用用户指定的值
// 	if req.CheckIntervalSeconds > 0 {
// 		autoConfig.CheckInterval = time.Duration(req.CheckIntervalSeconds) * time.Second
// 	}

// 	// 使用锁确保线程安全
// 	s.Engine.autoTradeLock.Lock()
// 	defer s.Engine.autoTradeLock.Unlock()

// 	// 如果已经有实例在运行，先停止它
// 	if s.Engine.autoTradeManager != nil {
// 		s.Engine.autoTradeManager.Stop()
// 	}

// 	// 创建新的自动交易管理器
// 	s.Engine.autoTradeManager = token.NewAutoTradeManager(s.Config, autoConfig)

// 	// 启动自动交易服务
// 	err = s.Engine.autoTradeManager.Start()
// 	if err != nil {
// 		return nil, fmt.Errorf("启动自动交易服务失败: %w", err)
// 	}

// 	return &gctrpc.StartAutoTradeResponse{
// 		Success: true,
// 		Message: "自动交易服务已成功启动",
// 	}, nil
// }

// // StopAutoTrade 停止自动交易服务
// func (s *RPCServer) StopAutoTrade(ctx context.Context, req *gctrpc.StopAutoTradeRequest) (*gctrpc.StopAutoTradeResponse, error) {
// 	// 使用锁确保线程安全
// 	s.Engine.autoTradeLock.Lock()
// 	defer s.Engine.autoTradeLock.Unlock()

// 	// 检查是否有正在运行的实例
// 	if s.Engine.autoTradeManager == nil {
// 		return &gctrpc.StopAutoTradeResponse{
// 			Success: false,
// 			Message: "没有正在运行的自动交易服务",
// 		}, nil
// 	}

// 	// 停止自动交易服务
// 	s.Engine.autoTradeManager.Stop()
// 	// 清除实例引用
// 	s.Engine.autoTradeManager = nil

// 	return &gctrpc.StopAutoTradeResponse{
// 		Success: true,
// 		Message: "自动交易服务已成功停止",
// 	}, nil
// }
