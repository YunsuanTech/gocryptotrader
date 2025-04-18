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
