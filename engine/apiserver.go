package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"gocryptotrader/common"
	"gocryptotrader/common/crypto"
	"gocryptotrader/config"
	"gocryptotrader/database/repository/xens"
	"gocryptotrader/exchanges/account"
	"gocryptotrader/exchanges/token"
	"gocryptotrader/log"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// setupAPIServerManager checks and creates an api server manager
func setupAPIServerManager(remoteConfig *config.RemoteControlConfig, pprofConfig *config.Profiler, portfolioManager iPortfolioManager) (*apiServerManager, error) {
	if remoteConfig == nil {
		return nil, errNilRemoteConfig
	}
	if pprofConfig == nil {
		return nil, errNilPProfConfig
	}

	cfg := config.GetConfig()
	return &apiServerManager{
		remoteConfig:           remoteConfig,
		pprofConfig:            pprofConfig,
		restListenAddress:      remoteConfig.DeprecatedRPC.ListenAddress,
		websocketListenAddress: remoteConfig.WebsocketRPC.ListenAddress,
		portfolioManager:       portfolioManager,
		accountManager:         account.New(cfg),
		watchlistManager:       token.NewWatchlistManager(cfg),
		tradingManager:         token.NewTradingManager(cfg),
	}, nil
}

// IsRESTServerRunning safely checks whether the subsystem is running
func (m *apiServerManager) IsRESTServerRunning() bool {
	if m == nil {
		return false
	}
	return atomic.LoadInt32(&m.restStarted) == 1
}

// IsWebsocketServerRunning safely checks whether the subsystem is running
func (m *apiServerManager) IsWebsocketServerRunning() bool {
	if m == nil {
		return false
	}
	return atomic.LoadInt32(&m.websocketStarted) == 1
}

// StopRESTServer attempts to shutdown the subsystem
func (m *apiServerManager) StopRESTServer() error {
	if m == nil {
		return fmt.Errorf("api server %w", ErrNilSubsystem)
	}
	if !atomic.CompareAndSwapInt32(&m.restStarted, 1, 0) {
		return fmt.Errorf("apiserver deprecated server %w", ErrSubSystemNotStarted)
	}
	err := m.restHTTPServer.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	m.wgRest.Wait()
	m.restRouter = nil
	return nil
}

func (m *apiServerManager) StopWebsocketServer() error {
	if m == nil {
		return fmt.Errorf("api server %w", ErrNilSubsystem)
	}
	if !atomic.CompareAndSwapInt32(&m.websocketStarted, 1, 0) {
		return fmt.Errorf("apiserver websocket server %w", ErrSubSystemNotStarted)
	}

	err := m.websocketHTTPServer.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	m.websocketRouter = nil
	m.websocketHub = nil
	m.wgWebsocket.Wait()
	m.websocketHTTPServer = nil
	return nil
}

// newRouter takes in the exchange interfaces and returns a new multiplexer
// router
func (m *apiServerManager) newRouter(isREST bool) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	var routes []Route
	if common.ExtractPort(m.websocketListenAddress) == 80 {
		m.websocketListenAddress = common.ExtractHost(m.websocketListenAddress)
	} else {
		m.websocketListenAddress = common.ExtractHost(m.websocketListenAddress) + ":" +
			strconv.Itoa(common.ExtractPort(m.websocketListenAddress))
	}

	if isREST {
		routes = []Route{
			{"", http.MethodGet, "/", m.getIndex},
			{"GetAllSettings", http.MethodGet, "/config/all", m.restGetAllSettings},
			{"SaveAllSettings", http.MethodPost, "/config/all/save", m.restSaveAllSettings},
			{"GetPortfolio", http.MethodGet, "/portfolio/all", m.restGetPortfolio},
			{"GetAccounts", http.MethodGet, "/accounts/all", m.restGetAccounts},
			{"GetWatchlistTokens", http.MethodGet, "/watchlist/tokens", m.restGetWatchlistTokens},
			{"GetWatchlistTokensByNetwork", http.MethodGet, "/watchlist/tokens/network/{network}", m.restGetWatchlistTokensByNetwork},
			{"GetWatchlistTokensBySymbol", http.MethodGet, "/watchlist/tokens/symbol/{symbol}", m.restGetWatchlistTokensBySymbol},
			{"AddWatchlistToken", http.MethodPost, "/watchlist/tokens/add", m.restAddWatchlistToken},
			{"UpdateWatchlistToken", http.MethodPut, "/watchlist/tokens/update/{id}", m.restUpdateWatchlistToken},
			{"UpdateWatchlistTokenByAddress", http.MethodPut, "/watchlist/tokens/update/address/{address}", m.restUpdateWatchlistTokenByAddress},
			{"DeleteWatchlistToken", http.MethodDelete, "/watchlist/tokens/delete/{id}", m.restDeleteWatchlistToken},
			{"DeleteWatchlistTokenByAddress", http.MethodDelete, "/watchlist/tokens/delete/address/{address}", m.restDeleteWatchlistTokenByAddress},
			// 添加交易规则相关路由
			{"GetAllTradingRules", http.MethodGet, "/trading/rules", m.restGetAllTradingRules},
			{"GetTradingRuleByID", http.MethodGet, "/trading/rules/{id}", m.restGetTradingRuleByID},
			{"GetTradingRulesByTokenID", http.MethodGet, "/trading/rules/token/{tokenId}", m.restGetTradingRulesByTokenID},
			{"GetTradingRulesByUserAddress", http.MethodGet, "/trading/rules/user/{userAddress}", m.restGetTradingRulesByUserAddress},
			{"GetTradingRulesByUserAndToken", http.MethodGet, "/trading/rules/user/{userAddress}/token/{tokenId}", m.restGetTradingRulesByUserAndToken},
			{"GetActiveTradingRules", http.MethodGet, "/trading/rules/active", m.restGetActiveTradingRules},
			{"AddTradingRule", http.MethodPost, "/trading/rules/add", m.restAddTradingRule},
			{"UpdateTradingRule", http.MethodPut, "/trading/rules/update/{id}", m.restUpdateTradingRule},
			{"DeleteTradingRule", http.MethodDelete, "/trading/rules/delete/{id}", m.restDeleteTradingRule},
			// 添加Xen相关路由
			{"GetXensByChainName", http.MethodGet, "/xens/chain/{chainName}", m.restGetXensByChainName},
			{"GetXensByStatusAndChain", http.MethodGet, "/xens/status/{status}/chain/{chainName}", m.restGetXensByStatusAndChain},
			{"AddXen", http.MethodPost, "/xens/add", m.restAddXen},
			{"UpdateXen", http.MethodPut, "/xens/update", m.restUpdateXen},
		}

		if m.pprofConfig.Enabled {
			if m.pprofConfig.MutexProfileFraction > 0 {
				runtime.SetMutexProfileFraction(m.pprofConfig.MutexProfileFraction)
			}
			log.Debugf(log.RESTSys,
				"HTTP Go performance profiler (pprof) endpoint enabled: http://%s:%d/debug/pprof/\n",
				common.ExtractHost(m.websocketListenAddress),
				common.ExtractPort(m.websocketListenAddress))
			router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
		}
	} else {
		routes = []Route{
			{"ws", http.MethodGet, "/ws", m.WebsocketClientHandler},
		}
	}

	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(restLogger(route.HandlerFunc, route.Name)).
			Host(m.websocketListenAddress)
	}
	return router
}

// StartRESTServer starts a REST handler
func (m *apiServerManager) StartRESTServer() error {
	if !atomic.CompareAndSwapInt32(&m.restStarted, 0, 1) {
		return fmt.Errorf("rest server %w", errAlreadyRunning)
	}
	if !m.remoteConfig.DeprecatedRPC.Enabled {
		atomic.StoreInt32(&m.restStarted, 0)
		return fmt.Errorf("rest %w", errServerDisabled)
	}
	log.Debugf(log.RESTSys,
		"Deprecated RPC handler support enabled. Listen URL: http://%s:%d\n",
		common.ExtractHost(m.restListenAddress), common.ExtractPort(m.restListenAddress))
	m.restRouter = m.newRouter(true)
	if m.restHTTPServer == nil {
		m.restHTTPServer = &http.Server{
			Addr:              m.restListenAddress,
			Handler:           m.restRouter,
			ReadHeaderTimeout: time.Minute,
		}
	}
	m.wgRest.Add(1)
	go func() {
		defer m.wgRest.Done()
		err := m.restHTTPServer.ListenAndServe()
		if err != nil {
			atomic.StoreInt32(&m.restStarted, 0)
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorln(log.APIServerMgr, err)
			}
		}
	}()
	return nil
}

// restLogger logs the requests internally
func restLogger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)

		log.Debugf(log.RESTSys,
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// writeResponse outputs a JSON response of the response interface
func writeResponse(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

// handleError prints the REST method and error
func handleError(method string, err error) {
	log.Errorf(log.APIServerMgr, "RESTful %s: handler failed to send JSON response. Error %s\n",
		method, err)
}

// restGetAllSettings replies to a request with an encoded JSON response about the
// trading Bots configuration.
func (m *apiServerManager) restGetAllSettings(w http.ResponseWriter, r *http.Request) {
	err := writeResponse(w, config.GetConfig())
	if err != nil {
		handleError(r.Method, err)
	}
}

// restSaveAllSettings saves all current settings from request body as a JSON
// document then reloads state and returns the settings
func (m *apiServerManager) restSaveAllSettings(w http.ResponseWriter, r *http.Request) {
	// Get the data from the request
	decoder := json.NewDecoder(r.Body)
	var responseData config.Post
	err := decoder.Decode(&responseData)
	if err != nil {
		handleError(r.Method, err)
	}
	// Save change the settings
	cfg := config.GetConfig()
	err = cfg.UpdateConfig(m.gctConfigPath, &responseData.Data, false)
	if err != nil {
		handleError(r.Method, err)
	}

	err = writeResponse(w, cfg)
	if err != nil {
		handleError(r.Method, err)
	}
	err = m.bot.SetupExchanges()
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetPortfolio returns the Bot portfolio manager
func (m *apiServerManager) restGetPortfolio(w http.ResponseWriter, r *http.Request) {
	result := m.portfolioManager.GetPortfolioSummary()
	err := writeResponse(w, result)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetAccounts 返回所有账户信息
func (m *apiServerManager) restGetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := m.accountManager.Accounts()
	if err != nil {
		handleError(r.Method, err)
		return
	}
	println(accounts)
	err = writeResponse(w, accounts)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetWatchlistTokens 返回所有代币监视列表
func (m *apiServerManager) restGetWatchlistTokens(w http.ResponseWriter, r *http.Request) {
	network := r.URL.Query().Get("network")
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // 默认限制
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	tokens, err := m.watchlistManager.GetAllWatchlistTokens(network, limit)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取代币监视列表失败: %s", err)
		http.Error(w, fmt.Sprintf("获取代币监视列表失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, tokens)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetWatchlistTokensByNetwork 根据网络返回代币监视列表
func (m *apiServerManager) restGetWatchlistTokensByNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	network := vars["network"]
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // 默认限制
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	tokens, err := m.watchlistManager.GetWatchlistTokensByNetwork(network, limit)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取网络代币监视列表失败: %s", err)
		http.Error(w, fmt.Sprintf("获取网络代币监视列表失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, tokens)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetWatchlistTokensBySymbol 根据符号返回代币监视列表
func (m *apiServerManager) restGetWatchlistTokensBySymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	tokens, err := m.watchlistManager.GetWatchlistTokensBySymbol(symbol)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取符号代币监视列表失败: %s", err)
		http.Error(w, fmt.Sprintf("获取符号代币监视列表失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, tokens)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restAddWatchlistToken 添加代币到监视列表
func (m *apiServerManager) restAddWatchlistToken(w http.ResponseWriter, r *http.Request) {
	type AddTokenRequest struct {
		TokenSymbol  string `json:"tokenSymbol"`
		TokenAddress string `json:"tokenAddress"`
		Network      string `json:"network"`
		Decimals     int    `json:"decimals"`
		IsActive     int    `json:"isActive"`
	}

	var req AddTokenRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析添加代币请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析添加代币请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.TokenSymbol == "" || req.TokenAddress == "" || req.Network == "" {
		http.Error(w, "代币符号、地址和网络为必填字段", http.StatusBadRequest)
		return
	}

	// 设置默认值
	creationTime := time.Now().Unix()
	lastUpdated := creationTime

	err = m.watchlistManager.AddWatchlistToken(
		req.TokenSymbol,
		req.TokenAddress,
		req.Network,
		req.Decimals,
		creationTime,
		lastUpdated,
		req.IsActive,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "添加代币到监视列表失败: %s", err)
		http.Error(w, fmt.Sprintf("添加代币到监视列表失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "代币已成功添加到监视列表"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restUpdateWatchlistToken 根据ID更新代币信息
func (m *apiServerManager) restUpdateWatchlistToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenIDStr := vars["id"]
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的代币ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的代币ID: %s", err), http.StatusBadRequest)
		return
	}

	type UpdateTokenRequest struct {
		TokenSymbol  string `json:"tokenSymbol"`
		TokenAddress string `json:"tokenAddress"`
		Network      string `json:"network"`
		Decimals     int    `json:"decimals"`
		IsActive     int    `json:"isActive"`
	}

	var req UpdateTokenRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析更新代币请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析更新代币请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.TokenSymbol == "" || req.TokenAddress == "" || req.Network == "" {
		http.Error(w, "代币符号、地址和网络为必填字段", http.StatusBadRequest)
		return
	}

	err = m.watchlistManager.UpdateWatchlistToken(
		tokenID,
		req.TokenSymbol,
		req.TokenAddress,
		req.Network,
		req.Decimals,
		req.IsActive,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "更新代币信息失败: %s", err)
		http.Error(w, fmt.Sprintf("更新代币信息失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "代币信息已成功更新"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restUpdateWatchlistTokenByAddress 根据地址更新代币信息
func (m *apiServerManager) restUpdateWatchlistTokenByAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["address"]

	type UpdateTokenRequest struct {
		TokenSymbol     string `json:"tokenSymbol"`
		NewTokenAddress string `json:"newTokenAddress"`
		Network         string `json:"network"`
		Decimals        int    `json:"decimals"`
		IsActive        int    `json:"isActive"`
	}

	var req UpdateTokenRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析更新代币请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析更新代币请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.TokenSymbol == "" || req.Network == "" {
		http.Error(w, "代币符号和网络为必填字段", http.StatusBadRequest)
		return
	}

	// 如果未提供新地址，则使用原地址
	newTokenAddress := req.NewTokenAddress
	if newTokenAddress == "" {
		newTokenAddress = tokenAddress
	}

	err = m.watchlistManager.UpdateWatchlistTokenByAddress(
		tokenAddress,
		req.TokenSymbol,
		newTokenAddress,
		req.Network,
		req.Decimals,
		req.IsActive,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "更新代币信息失败: %s", err)
		http.Error(w, fmt.Sprintf("更新代币信息失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "代币信息已成功更新"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restDeleteWatchlistToken 根据ID删除代币
func (m *apiServerManager) restDeleteWatchlistToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenIDStr := vars["id"]
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的代币ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的代币ID: %s", err), http.StatusBadRequest)
		return
	}

	err = m.watchlistManager.DeleteWatchlistToken(tokenID)
	if err != nil {
		log.Errorf(log.APIServerMgr, "删除代币失败: %s", err)
		http.Error(w, fmt.Sprintf("删除代币失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "代币已成功删除"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restDeleteWatchlistTokenByAddress 根据地址删除代币
func (m *apiServerManager) restDeleteWatchlistTokenByAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["address"]

	err := m.watchlistManager.DeleteWatchlistTokenByAddress(tokenAddress)
	if err != nil {
		log.Errorf(log.APIServerMgr, "删除代币失败: %s", err)
		http.Error(w, fmt.Sprintf("删除代币失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "代币已成功删除"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetXensByChainName 根据链名获取Xen记录
func (m *apiServerManager) restGetXensByChainName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chainName := vars["chainName"]

	xenRecords, err := xens.GetXensByChainName(chainName)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取Xen记录失败: %s", err)
		http.Error(w, fmt.Sprintf("获取Xen记录失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, xenRecords)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetXensByStatusAndChain 根据状态和链名获取Xen记录
func (m *apiServerManager) restGetXensByStatusAndChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status := vars["status"]
	chainName := vars["chainName"]

	xenRecords, err := xens.GetXensByStatusByChainName(status, chainName)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取Xen记录失败: %s", err)
		http.Error(w, fmt.Sprintf("获取Xen记录失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, xenRecords)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restAddXen 添加新的Xen记录
func (m1 *apiServerManager) restAddXen(w http.ResponseWriter, r *http.Request) {
	type AddXenRequest struct {
		Slot           int      `json:"slot"`
		ChainName      string   `json:"chainName"`
		Count          int      `json:"count"`
		Days           int      `json:"days"`
		ExecutionTime  *string  `json:"executionTime,omitempty"`
		ClaimTime      *string  `json:"claimTime,omitempty"`
		ExpectedReward *float64 `json:"expectedReward,omitempty"`
		Ranking        int      `json:"ranking"`
		Amp            int      `json:"amp"`
		Eaa            float64  `json:"eaa"`
		M              *float64 `json:"m,omitempty"`
		Status         string   `json:"status"`
		TxID           *string  `json:"txId,omitempty"`
		MintFees       *float64 `json:"mintFees,omitempty"`
		ClaimFees      *float64 `json:"claimFees,omitempty"`
	}

	var req AddXenRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析添加Xen请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析添加Xen请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.ChainName == "" || req.Status == "" {
		http.Error(w, "链名和状态为必填字段", http.StatusBadRequest)
		return
	}

	// 处理可选字段
	var executionTime sql.NullTime
	if req.ExecutionTime != nil {
		t, err := time.Parse(time.RFC3339, *req.ExecutionTime)
		if err == nil {
			executionTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	var claimTime sql.NullTime
	if req.ClaimTime != nil {
		t, err := time.Parse(time.RFC3339, *req.ClaimTime)
		if err == nil {
			claimTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	var expectedReward sql.NullFloat64
	if req.ExpectedReward != nil {
		expectedReward = sql.NullFloat64{Float64: *req.ExpectedReward, Valid: true}
	}

	var m sql.NullFloat64
	if req.M != nil {
		m = sql.NullFloat64{Float64: *req.M, Valid: true}
	}

	var txID sql.NullString
	if req.TxID != nil {
		txID = sql.NullString{String: *req.TxID, Valid: true}
	}

	var mintFees sql.NullFloat64
	if req.MintFees != nil {
		mintFees = sql.NullFloat64{Float64: *req.MintFees, Valid: true}
	}

	var claimFees sql.NullFloat64
	if req.ClaimFees != nil {
		claimFees = sql.NullFloat64{Float64: *req.ClaimFees, Valid: true}
	}

	err = xens.AddXen(
		req.Slot,
		req.ChainName,
		req.Count,
		req.Days,
		executionTime,
		claimTime,
		expectedReward,
		req.Ranking,
		req.Amp,
		req.Eaa,
		m,
		req.Status,
		txID,
		mintFees,
		claimFees,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "添加Xen记录失败: %s", err)
		http.Error(w, fmt.Sprintf("添加Xen记录失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "Xen记录已成功添加"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restUpdateXen 更新Xen记录
func (m1 *apiServerManager) restUpdateXen(w http.ResponseWriter, r *http.Request) {
	type UpdateXenRequest struct {
		Slot           int      `json:"slot"`
		ChainName      string   `json:"chainName"`
		Count          int      `json:"count"`
		Days           int      `json:"days"`
		ExecutionTime  *string  `json:"executionTime,omitempty"`
		ClaimTime      *string  `json:"claimTime,omitempty"`
		ExpectedReward *float64 `json:"expectedReward,omitempty"`
		Ranking        int      `json:"ranking"`
		Amp            int      `json:"amp"`
		Eaa            float64  `json:"eaa"`
		M              *float64 `json:"m,omitempty"`
		Status         string   `json:"status"`
		TxID           *string  `json:"txId,omitempty"`
		MintFees       *float64 `json:"mintFees,omitempty"`
		ClaimFees      *float64 `json:"claimFees,omitempty"`
	}

	var req UpdateXenRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析更新Xen请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析更新Xen请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.ChainName == "" || req.Status == "" {
		http.Error(w, "链名和状态为必填字段", http.StatusBadRequest)
		return
	}

	// 处理可选字段
	var executionTime sql.NullTime
	if req.ExecutionTime != nil {
		t, err := time.Parse(time.RFC3339, *req.ExecutionTime)
		if err == nil {
			executionTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	var claimTime sql.NullTime
	if req.ClaimTime != nil {
		t, err := time.Parse(time.RFC3339, *req.ClaimTime)
		if err == nil {
			claimTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	var expectedReward sql.NullFloat64
	if req.ExpectedReward != nil {
		expectedReward = sql.NullFloat64{Float64: *req.ExpectedReward, Valid: true}
	}

	var m sql.NullFloat64
	if req.M != nil {
		m = sql.NullFloat64{Float64: *req.M, Valid: true}
	}

	var txID sql.NullString
	if req.TxID != nil {
		txID = sql.NullString{String: *req.TxID, Valid: true}
	}

	var mintFees sql.NullFloat64
	if req.MintFees != nil {
		mintFees = sql.NullFloat64{Float64: *req.MintFees, Valid: true}
	}

	var claimFees sql.NullFloat64
	if req.ClaimFees != nil {
		claimFees = sql.NullFloat64{Float64: *req.ClaimFees, Valid: true}
	}

	err = xens.UpdateXen(
		req.Slot,
		req.ChainName,
		req.Count,
		req.Days,
		executionTime,
		claimTime,
		expectedReward,
		req.Ranking,
		req.Amp,
		req.Eaa,
		m,
		req.Status,
		txID,
		mintFees,
		claimFees,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "更新Xen记录失败: %s", err)
		http.Error(w, fmt.Sprintf("更新Xen记录失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "Xen记录已成功更新"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// wsGetAccounts 处理WebSocket的getaccounts请求
func wsGetAccounts(client *websocketClient, _ interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "getaccounts",
	}

	// 使用accountManager获取账户信息
	accountManager := account.New(config.GetConfig())
	accounts, err := accountManager.Accounts()
	if err != nil {
		wsResp.Error = err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	log.Debugf(log.APIServerMgr, "websocket: sending accounts data, count: %d", len(accounts))
	wsResp.Data = accounts
	return client.SendWebsocketMessage(wsResp)
}

// wsGetWatchlistTokens 处理WebSocket的getwatchlisttokens请求
func wsGetWatchlistTokens(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "getwatchlisttokens",
	}

	// 解析请求参数
	type GetTokensRequest struct {
		Network string `json:"network"`
		Limit   int    `json:"limit"`
	}

	var req GetTokensRequest
	req.Limit = 100 // 默认限制

	if data != nil {
		d, ok := data.([]byte)
		if ok {
			err := json.Unmarshal(d, &req)
			if err != nil {
				log.Debugf(log.APIServerMgr, "websocket: failed to parse getwatchlisttokens request: %s", err)
			}
		}
	}

	// 使用watchlistManager获取代币信息
	tokens, err := client.watchlistManager.GetAllWatchlistTokens(req.Network, req.Limit)
	if err != nil {
		wsResp.Error = err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	log.Debugf(log.APIServerMgr, "websocket: sending watchlist tokens data, count: %d", len(tokens))
	wsResp.Data = tokens
	return client.SendWebsocketMessage(wsResp)
}

// wsGetXensByChainName 处理WebSocket的getxensbychainname请求
func wsGetXensByChainName(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "getxensbychainname",
	}

	// 解析请求参数
	type GetXensRequest struct {
		ChainName string `json:"chainName"`
	}

	var req GetXensRequest
	if data != nil {
		d, ok := data.([]byte)
		if ok {
			err := json.Unmarshal(d, &req)
			if err != nil {
				log.Debugf(log.APIServerMgr, "websocket: failed to parse getxensbychainname request: %s", err)
			}
		}
	}

	// 验证必填字段
	if req.ChainName == "" {
		wsResp.Error = "链名为必填字段"
		return client.SendWebsocketMessage(wsResp)
	}

	// 获取Xen记录
	xenRecords, err := xens.GetXensByChainName(req.ChainName)
	if err != nil {
		wsResp.Error = err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	log.Debugf(log.APIServerMgr, "websocket: sending xens data for chain %s", req.ChainName)
	wsResp.Data = xenRecords
	return client.SendWebsocketMessage(wsResp)
}

// wsGetXensByStatusAndChain 处理WebSocket的getxensbystatusandchain请求
func wsGetXensByStatusAndChain(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "getxensbystatusandchain",
	}

	// 解析请求参数
	type GetXensStatusRequest struct {
		Status    string `json:"status"`
		ChainName string `json:"chainName"`
	}

	var req GetXensStatusRequest
	if data != nil {
		d, ok := data.([]byte)
		if ok {
			err := json.Unmarshal(d, &req)
			if err != nil {
				log.Debugf(log.APIServerMgr, "websocket: failed to parse getxensbystatusandchain request: %s", err)
			}
		}
	}

	// 验证必填字段
	if req.Status == "" || req.ChainName == "" {
		wsResp.Error = "状态和链名为必填字段"
		return client.SendWebsocketMessage(wsResp)
	}

	// 获取Xen记录
	xenRecords, err := xens.GetXensByStatusByChainName(req.Status, req.ChainName)
	if err != nil {
		wsResp.Error = err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	log.Debugf(log.APIServerMgr, "websocket: sending xens data for status %s and chain %s", req.Status, req.ChainName)
	wsResp.Data = xenRecords
	return client.SendWebsocketMessage(wsResp)
}

// wsAddWatchlistToken 处理WebSocket的addwatchlisttoken请求
func wsAddWatchlistToken(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "addwatchlisttoken",
	}

	// 解析请求参数
	type AddTokenRequest struct {
		TokenSymbol  string `json:"tokenSymbol"`
		TokenAddress string `json:"tokenAddress"`
		Network      string `json:"network"`
		Decimals     int    `json:"decimals"`
		IsActive     int    `json:"isActive"`
	}

	var req AddTokenRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.TokenSymbol == "" || req.TokenAddress == "" || req.Network == "" {
		wsResp.Error = "tokenSymbol, tokenAddress and network are required fields"
		return client.SendWebsocketMessage(wsResp)
	}

	// 设置默认值
	creationTime := time.Now().Unix()
	lastUpdated := creationTime
	isActive := req.IsActive
	if isActive == 0 {
		isActive = 1 // 默认为活跃状态
	}

	err = client.watchlistManager.AddWatchlistToken(
		req.TokenSymbol,
		req.TokenAddress,
		req.Network,
		req.Decimals,
		creationTime,
		lastUpdated,
		isActive,
	)

	if err != nil {
		wsResp.Error = "failed to add token to watchlist: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "token successfully added to watchlist"}
	return client.SendWebsocketMessage(wsResp)
}

// wsUpdateWatchlistToken 处理WebSocket的updatewatchlisttoken请求
func wsUpdateWatchlistToken(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "updatewatchlisttoken",
	}

	// 解析请求参数
	type UpdateTokenRequest struct {
		TokenID      int    `json:"tokenId"`
		TokenSymbol  string `json:"tokenSymbol"`
		TokenAddress string `json:"tokenAddress"`
		Network      string `json:"network"`
		Decimals     int    `json:"decimals"`
		IsActive     int    `json:"isActive"`
	}

	var req UpdateTokenRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.TokenID <= 0 {
		wsResp.Error = "tokenId is required and must be greater than 0"
		return client.SendWebsocketMessage(wsResp)
	}

	if req.TokenSymbol == "" || req.TokenAddress == "" || req.Network == "" {
		wsResp.Error = "tokenSymbol, tokenAddress and network are required fields"
		return client.SendWebsocketMessage(wsResp)
	}

	// 设置默认值
	isActive := req.IsActive
	if isActive == 0 {
		isActive = 1 // 默认为活跃状态
	}

	err = client.watchlistManager.UpdateWatchlistToken(
		req.TokenID,
		req.TokenSymbol,
		req.TokenAddress,
		req.Network,
		req.Decimals,
		isActive,
	)

	if err != nil {
		wsResp.Error = "failed to update token: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "token successfully updated"}
	return client.SendWebsocketMessage(wsResp)
}

// wsUpdateWatchlistTokenByAddress 处理WebSocket的updatewatchlisttokenbyaddress请求
func wsUpdateWatchlistTokenByAddress(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "updatewatchlisttokenbyaddress",
	}

	// 解析请求参数
	type UpdateTokenRequest struct {
		TokenAddress    string `json:"tokenAddress"`
		TokenSymbol     string `json:"tokenSymbol"`
		NewTokenAddress string `json:"newTokenAddress"`
		Network         string `json:"network"`
		Decimals        int    `json:"decimals"`
		IsActive        int    `json:"isActive"`
	}

	var req UpdateTokenRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.TokenAddress == "" {
		wsResp.Error = "tokenAddress is required"
		return client.SendWebsocketMessage(wsResp)
	}

	if req.TokenSymbol == "" || req.Network == "" {
		wsResp.Error = "tokenSymbol and network are required fields"
		return client.SendWebsocketMessage(wsResp)
	}

	// 设置默认值
	isActive := req.IsActive
	if isActive == 0 {
		isActive = 1 // 默认为活跃状态
	}

	// 如果未提供新地址，则使用原地址
	newTokenAddress := req.NewTokenAddress
	if newTokenAddress == "" {
		newTokenAddress = req.TokenAddress
	}

	err = client.watchlistManager.UpdateWatchlistTokenByAddress(
		req.TokenAddress,
		req.TokenSymbol,
		newTokenAddress,
		req.Network,
		req.Decimals,
		isActive,
	)

	if err != nil {
		wsResp.Error = "failed to update token: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "token successfully updated"}
	return client.SendWebsocketMessage(wsResp)
}

// wsDeleteWatchlistToken 处理WebSocket的deletewatchlisttoken请求
func wsDeleteWatchlistToken(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "deletewatchlisttoken",
	}

	// 解析请求参数
	type DeleteTokenRequest struct {
		TokenID int `json:"tokenId"`
	}

	var req DeleteTokenRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.TokenID <= 0 {
		wsResp.Error = "tokenId is required and must be greater than 0"
		return client.SendWebsocketMessage(wsResp)
	}

	err = client.watchlistManager.DeleteWatchlistToken(req.TokenID)
	if err != nil {
		wsResp.Error = "failed to delete token: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "token successfully deleted"}
	return client.SendWebsocketMessage(wsResp)
}

// wsDeleteWatchlistTokenByAddress 处理WebSocket的deletewatchlisttokenbyaddress请求
func wsDeleteWatchlistTokenByAddress(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "deletewatchlisttokenbyaddress",
	}

	// 解析请求参数
	type DeleteTokenRequest struct {
		TokenAddress string `json:"tokenAddress"`
	}

	var req DeleteTokenRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.TokenAddress == "" {
		wsResp.Error = "tokenAddress is required"
		return client.SendWebsocketMessage(wsResp)
	}

	err = client.watchlistManager.DeleteWatchlistTokenByAddress(req.TokenAddress)
	if err != nil {
		wsResp.Error = "failed to delete token: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "token successfully deleted"}
	return client.SendWebsocketMessage(wsResp)
}

// getIndex returns an HTML snippet for when a user requests the index URL
func (m *apiServerManager) getIndex(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(w, restIndexResponse)
	if err != nil {
		log.Errorln(log.APIServerMgr, err)
	}
	w.WriteHeader(http.StatusOK)
}

// StartWebsocketServer starts a Websocket handler
func (m *apiServerManager) StartWebsocketServer() error {
	if !atomic.CompareAndSwapInt32(&m.websocketStarted, 0, 1) {
		return fmt.Errorf("websocket server %w", errAlreadyRunning)
	}
	if !m.remoteConfig.WebsocketRPC.Enabled {
		atomic.StoreInt32(&m.websocketStarted, 0)
		return fmt.Errorf("websocket %w", errServerDisabled)
	}
	log.Debugf(log.APIServerMgr,
		"Websocket RPC support enabled. Listen URL: ws://%s:%d/ws\n",
		common.ExtractHost(m.websocketListenAddress), common.ExtractPort(m.websocketListenAddress))
	m.websocketRouter = m.newRouter(false)
	if m.websocketHTTPServer == nil {
		m.websocketHTTPServer = &http.Server{
			Addr:              m.websocketListenAddress,
			Handler:           m.websocketRouter,
			ReadHeaderTimeout: time.Minute,
		}
	}

	m.wgWebsocket.Add(1)
	go func() {
		defer m.wgWebsocket.Done()
		err := m.websocketHTTPServer.ListenAndServe()
		if err != nil {
			atomic.StoreInt32(&m.websocketStarted, 0)
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorln(log.APIServerMgr, err)
			}
		}
	}()
	return nil
}

// newWebsocketHub Creates a new websocket hub
func newWebsocketHub() *websocketHub {
	return &websocketHub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocketClient),
		Unregister: make(chan *websocketClient),
		Clients:    make(map[*websocketClient]bool),
	}
}

func (h *websocketHub) run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				log.Debugln(log.APIServerMgr, "websocket: disconnected client")
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					log.Debugln(log.APIServerMgr, "websocket: disconnected client")
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// SendWebsocketMessage sends a websocket event to the client
func (c *websocketClient) SendWebsocketMessage(evt interface{}) error {
	data, err := json.Marshal(evt)
	if err != nil {
		log.Errorf(log.APIServerMgr, "websocket: failed to send message: %s\n", err)
		return err
	}

	c.Send <- data
	return nil
}

func (c *websocketClient) read() {
	defer func() {
		c.Hub.Unregister <- c
		conErr := c.Conn.Close()
		if conErr != nil {
			log.Errorln(log.APIServerMgr, conErr)
		}
	}()

	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf(log.APIServerMgr, "websocket: client disconnected, err: %s\n", err)
			}
			break
		}

		if msgType == websocket.TextMessage {
			var evt WebsocketEvent
			err := json.Unmarshal(message, &evt)
			if err != nil {
				log.Errorf(log.APIServerMgr, "websocket: failed to decode JSON sent from client %s\n", err)
				continue
			}

			dataJSON, err := json.Marshal(evt.Data)
			if err != nil {
				log.Errorln(log.APIServerMgr, "websocket: client sent data we couldn't JSON decode")
				break
			}

			req := strings.ToLower(evt.Event)
			log.Debugf(log.APIServerMgr, "websocket: request received: %s\n", req)

			result, ok := wsHandlers[req]
			if !ok {
				log.Debugln(log.APIServerMgr, "websocket: unsupported event")
				continue
			}

			if result.authRequired && !c.Authenticated {
				log.Warnf(log.APIServerMgr, "Websocket: request %s failed due to unauthenticated request on an authenticated API\n", evt.Event)
				err = c.SendWebsocketMessage(WebsocketEventResponse{Event: evt.Event, Error: "unauthorised request on authenticated API"})
				if err != nil {
					log.Errorln(log.APIServerMgr, err)
				}
				continue
			}

			err = result.handler(c, dataJSON)
			if err != nil {
				log.Errorf(log.APIServerMgr, "websocket: request %s failed. Error %s\n", evt.Event, err)
				continue
			}
		}
	}
}

func (c *websocketClient) write() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			log.Errorln(log.APIServerMgr, err)
		}
	}()
	for {
		message, ok := <-c.Send
		if !ok {
			err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			if err != nil {
				log.Errorln(log.APIServerMgr, err)
			}
			log.Debugln(log.APIServerMgr, "websocket: hub closed the channel")
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Errorf(log.APIServerMgr, "websocket: failed to create new io.writeCloser: %s\n", err)
			return
		}
		_, err = w.Write(message)
		if err != nil {
			log.Errorln(log.APIServerMgr, err)
		}

		// Add queued chat messages to the current websocket message
		n := len(c.Send)
		for range n {
			_, err = w.Write(<-c.Send)
			if err != nil {
				log.Errorln(log.APIServerMgr, err)
			}
		}

		if err := w.Close(); err != nil {
			log.Errorf(log.APIServerMgr, "websocket: failed to close io.WriteCloser: %s\n", err)
			return
		}
	}
}

// StartWebsocketHandler starts the websocket hub and routine which
// handles clients
func StartWebsocketHandler() {
	if !wsHubStarted {
		wsHubStarted = true
		wsHub = newWebsocketHub()
		go wsHub.run()
	}
}

// BroadcastWebsocketMessage meow
func BroadcastWebsocketMessage(evt WebsocketEvent) error {
	if !wsHubStarted {
		return ErrWebsocketServiceNotRunning
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	wsHub.Broadcast <- data
	return nil
}

// WebsocketClientHandler upgrades the HTTP connection to a websocket
// compatible one
func (m *apiServerManager) WebsocketClientHandler(w http.ResponseWriter, r *http.Request) {
	if !wsHubStarted {
		StartWebsocketHandler()
	}

	connectionLimit := m.remoteConfig.WebsocketRPC.ConnectionLimit
	numClients := len(wsHub.Clients)

	if numClients >= connectionLimit {
		log.Warnf(log.APIServerMgr,
			"websocket: client rejected due to websocket client limit reached. Number of clients %d. Limit %d.\n",
			numClients, connectionLimit)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	upgrader := websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}

	// Allow insecure origin if the Origin request header is present and not
	// equal to the Host request header. Default to false
	if m.remoteConfig.WebsocketRPC.AllowInsecureOrigin {
		upgrader.CheckOrigin = func(*http.Request) bool { return true }
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln(log.APIServerMgr, err)
		return
	}

	client := &websocketClient{
		Hub:              wsHub,
		Conn:             conn,
		Send:             make(chan []byte, 1024),
		maxAuthFailures:  m.remoteConfig.WebsocketRPC.MaxAuthFailures,
		username:         m.remoteConfig.Username,
		password:         m.remoteConfig.Password,
		configPath:       m.gctConfigPath,
		bot:              m.bot,
		portfolioManager: m.portfolioManager,
		watchlistManager: m.watchlistManager,
	}

	client.Hub.Register <- client
	log.Debugf(log.APIServerMgr,
		"websocket: client connected. Connected clients: %d. Limit %d.\n",
		numClients+1, connectionLimit)

	go client.read()
	go client.write()
}

func wsAuth(client *websocketClient, data interface{}) error {
	d, ok := data.([]byte)
	if !ok {
		return common.GetTypeAssertError("[]byte", data)
	}

	wsResp := WebsocketEventResponse{
		Event: "auth",
	}

	var auth WebsocketAuth
	err := json.Unmarshal(d, &auth)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Errorln(log.APIServerMgr, sendErr)
		}
		return err
	}

	hash, err := crypto.GetSHA256([]byte(client.password))

	if err != nil {
		return err
	}

	hashPW := crypto.HexEncodeToString(hash)
	if auth.Username == client.username && auth.Password == hashPW {
		client.Authenticated = true
		wsResp.Data = WebsocketResponseSuccess
		log.Debugln(log.APIServerMgr,
			"websocket: client authenticated successfully")
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Error = "invalid username/password"
	client.authFailures++
	sendErr := client.SendWebsocketMessage(wsResp)
	if sendErr != nil {
		log.Errorln(log.APIServerMgr, sendErr)
	}
	if client.authFailures >= client.maxAuthFailures {
		log.Debugf(log.APIServerMgr,
			"websocket: disconnecting client, maximum auth failures threshold reached (failures: %d limit: %d)\n",
			client.authFailures, client.maxAuthFailures)
		wsHub.Unregister <- client
		return nil
	}

	log.Debugf(log.APIServerMgr,
		"websocket: client sent wrong username/password (failures: %d limit: %d)\n",
		client.authFailures, client.maxAuthFailures)
	return nil
}

// restGetAllTradingRules 返回所有交易规则
func (m *apiServerManager) restGetAllTradingRules(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // 默认限制
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	rules, err := m.tradingManager.GetAllTradingRules(limit)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取交易规则列表失败: %s", err)
		http.Error(w, fmt.Sprintf("获取交易规则列表失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, rules)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetTradingRuleByID 根据ID获取交易规则
func (m *apiServerManager) restGetTradingRuleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleIDStr := vars["id"]
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的规则ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的规则ID: %s", err), http.StatusBadRequest)
		return
	}

	rule, err := m.tradingManager.GetTradingRuleByID(ruleID)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("获取交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, rule)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetTradingRulesByTokenID 根据代币ID获取交易规则
func (m *apiServerManager) restGetTradingRulesByTokenID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenIDStr := vars["tokenId"]
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的代币ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的代币ID: %s", err), http.StatusBadRequest)
		return
	}

	rules, err := m.tradingManager.GetTradingRulesByTokenID(tokenID)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取代币交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("获取代币交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, rules)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetTradingRulesByUserAddress 根据用户地址获取交易规则
func (m *apiServerManager) restGetTradingRulesByUserAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userAddress := vars["userAddress"]

	rules, err := m.tradingManager.GetTradingRulesByUserAddress(userAddress)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取用户交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("获取用户交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, rules)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetTradingRulesByUserAndToken 根据用户地址和代币ID获取交易规则
func (m *apiServerManager) restGetTradingRulesByUserAndToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userAddress := vars["userAddress"]
	tokenIDStr := vars["tokenId"]
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的代币ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的代币ID: %s", err), http.StatusBadRequest)
		return
	}

	rules, err := m.tradingManager.GetTradingRulesByUserAndToken(userAddress, tokenID)
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取用户代币交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("获取用户代币交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, rules)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restGetActiveTradingRules 获取所有活跃的交易规则
func (m *apiServerManager) restGetActiveTradingRules(w http.ResponseWriter, r *http.Request) {
	rules, err := m.tradingManager.GetActiveTradingRules()
	if err != nil {
		log.Errorf(log.APIServerMgr, "获取活跃交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("获取活跃交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	err = writeResponse(w, rules)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restAddTradingRule 添加新的交易规则
func (m *apiServerManager) restAddTradingRule(w http.ResponseWriter, r *http.Request) {
	type AddRuleRequest struct {
		TokenID        int     `json:"tokenId"`
		UserAddress    string  `json:"userAddress"`
		Direction      string  `json:"direction"`
		TriggerPrice   float64 `json:"triggerPrice"`
		Quantity       float64 `json:"quantity"`
		Slippage       float64 `json:"slippage"`
		ExpirationTime int64   `json:"expirationTime"`
		IsEnabled      int     `json:"isEnabled"`
		OrderType      string  `json:"orderType"`
	}

	var req AddRuleRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析添加交易规则请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析添加交易规则请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.UserAddress == "" || req.Direction == "" || req.OrderType == "" {
		http.Error(w, "用户地址、交易方向和订单类型为必填字段", http.StatusBadRequest)
		return
	}

	err = m.tradingManager.AddTradingRule(
		req.TokenID,
		req.UserAddress,
		req.Direction,
		req.TriggerPrice,
		req.Quantity,
		req.Slippage,
		req.ExpirationTime,
		req.IsEnabled,
		req.OrderType,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "添加交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("添加交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "交易规则已成功添加"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restUpdateTradingRule 更新交易规则
func (m *apiServerManager) restUpdateTradingRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleIDStr := vars["id"]
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的规则ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的规则ID: %s", err), http.StatusBadRequest)
		return
	}

	type UpdateRuleRequest struct {
		TokenID        int     `json:"tokenId"`
		UserAddress    string  `json:"userAddress"`
		Direction      string  `json:"direction"`
		TriggerPrice   float64 `json:"triggerPrice"`
		Quantity       float64 `json:"quantity"`
		Slippage       float64 `json:"slippage"`
		ExpirationTime int64   `json:"expirationTime"`
		IsEnabled      int     `json:"isEnabled"`
		OrderType      string  `json:"orderType"`
	}

	var req UpdateRuleRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Errorf(log.APIServerMgr, "解析更新交易规则请求失败: %s", err)
		http.Error(w, fmt.Sprintf("解析更新交易规则请求失败: %s", err), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.UserAddress == "" || req.Direction == "" || req.OrderType == "" {
		http.Error(w, "用户地址、交易方向和订单类型为必填字段", http.StatusBadRequest)
		return
	}

	err = m.tradingManager.UpdateTradingRule(
		ruleID,
		req.TokenID,
		req.UserAddress,
		req.Direction,
		req.TriggerPrice,
		req.Quantity,
		req.Slippage,
		req.ExpirationTime,
		req.IsEnabled,
		req.OrderType,
	)

	if err != nil {
		log.Errorf(log.APIServerMgr, "更新交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("更新交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "交易规则已成功更新"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// restDeleteTradingRule 删除交易规则
func (m *apiServerManager) restDeleteTradingRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleIDStr := vars["id"]
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		log.Errorf(log.APIServerMgr, "无效的规则ID: %s", err)
		http.Error(w, fmt.Sprintf("无效的规则ID: %s", err), http.StatusBadRequest)
		return
	}

	err = m.tradingManager.DeleteTradingRule(ruleID)
	if err != nil {
		log.Errorf(log.APIServerMgr, "删除交易规则失败: %s", err)
		http.Error(w, fmt.Sprintf("删除交易规则失败: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "success", "message": "交易规则已成功删除"}
	err = writeResponse(w, response)
	if err != nil {
		handleError(r.Method, err)
	}
}

// wsGetTradingRules 处理WebSocket的gettradingrules请求
func wsGetTradingRules(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "gettradingrules",
	}

	// 解析请求参数
	type GetRulesRequest struct {
		Limit int `json:"limit"`
	}

	var req GetRulesRequest
	req.Limit = 100 // 默认限制

	if data != nil {
		d, ok := data.([]byte)
		if ok {
			err := json.Unmarshal(d, &req)
			if err != nil {
				log.Debugf(log.APIServerMgr, "websocket: failed to parse gettradingrules request: %s", err)
			}
		}
	}

	// 使用tradingManager获取交易规则
	rules, err := client.tradingManager.GetAllTradingRules(req.Limit)
	if err != nil {
		wsResp.Error = err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	log.Debugf(log.APIServerMgr, "websocket: sending trading rules data, count: %d", len(rules))
	wsResp.Data = rules
	return client.SendWebsocketMessage(wsResp)
}

// wsAddTradingRule 处理WebSocket的addtradingrule请求
func wsAddTradingRule(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "addtradingrule",
	}

	// 解析请求参数
	type AddRuleRequest struct {
		TokenID        int     `json:"tokenId"`
		UserAddress    string  `json:"userAddress"`
		Direction      string  `json:"direction"`
		TriggerPrice   float64 `json:"triggerPrice"`
		Quantity       float64 `json:"quantity"`
		Slippage       float64 `json:"slippage"`
		ExpirationTime int64   `json:"expirationTime"`
		IsEnabled      int     `json:"isEnabled"`
		OrderType      string  `json:"orderType"`
	}

	var req AddRuleRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.UserAddress == "" || req.Direction == "" || req.OrderType == "" {
		wsResp.Error = "userAddress, direction and orderType are required fields"
		return client.SendWebsocketMessage(wsResp)
	}

	err = client.tradingManager.AddTradingRule(
		req.TokenID,
		req.UserAddress,
		req.Direction,
		req.TriggerPrice,
		req.Quantity,
		req.Slippage,
		req.ExpirationTime,
		req.IsEnabled,
		req.OrderType,
	)

	if err != nil {
		wsResp.Error = "failed to add trading rule: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "trading rule successfully added"}
	return client.SendWebsocketMessage(wsResp)
}

// wsUpdateTradingRule 处理WebSocket的updatetradingrule请求
func wsUpdateTradingRule(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "updatetradingrule",
	}

	// 解析请求参数
	type UpdateRuleRequest struct {
		RuleID         int     `json:"ruleId"`
		TokenID        int     `json:"tokenId"`
		UserAddress    string  `json:"userAddress"`
		Direction      string  `json:"direction"`
		TriggerPrice   float64 `json:"triggerPrice"`
		Quantity       float64 `json:"quantity"`
		Slippage       float64 `json:"slippage"`
		ExpirationTime int64   `json:"expirationTime"`
		IsEnabled      int     `json:"isEnabled"`
		OrderType      string  `json:"orderType"`
	}

	var req UpdateRuleRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	// 验证必填字段
	if req.UserAddress == "" || req.Direction == "" || req.OrderType == "" {
		wsResp.Error = "userAddress, direction and orderType are required fields"
		return client.SendWebsocketMessage(wsResp)
	}

	err = client.tradingManager.UpdateTradingRule(
		req.RuleID,
		req.TokenID,
		req.UserAddress,
		req.Direction,
		req.TriggerPrice,
		req.Quantity,
		req.Slippage,
		req.ExpirationTime,
		req.IsEnabled,
		req.OrderType,
	)

	if err != nil {
		wsResp.Error = "failed to update trading rule: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "trading rule successfully updated"}
	return client.SendWebsocketMessage(wsResp)
}

// wsDeleteTradingRule 处理WebSocket的deletetradingrule请求
func wsDeleteTradingRule(client *websocketClient, data interface{}) error {
	if client == nil {
		return errors.New("client is nil")
	}

	wsResp := WebsocketEventResponse{
		Event: "deletetradingrule",
	}

	// 解析请求参数
	type DeleteRuleRequest struct {
		RuleID int `json:"ruleId"`
	}

	var req DeleteRuleRequest
	d, ok := data.([]byte)
	if !ok {
		wsResp.Error = "invalid request data"
		return client.SendWebsocketMessage(wsResp)
	}

	err := json.Unmarshal(d, &req)
	if err != nil {
		wsResp.Error = "failed to parse request: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	err = client.tradingManager.DeleteTradingRule(req.RuleID)
	if err != nil {
		wsResp.Error = "failed to delete trading rule: " + err.Error()
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Data = map[string]string{"status": "success", "message": "trading rule successfully deleted"}
	return client.SendWebsocketMessage(wsResp)
}
