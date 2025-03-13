package engine

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"gocryptotrader/config"
	"gocryptotrader/database"
	gctlog "gocryptotrader/log"
	"gocryptotrader/utils"
)

// Engine contains configuration, portfolio manager, exchange & ticker data and is the
// overarching type across this code base.
type Engine struct {
	Config          *config.Config
	DatabaseManager *DatabaseConnectionManager
	Settings        Settings
	ServicesWG      sync.WaitGroup
}

// Bot is a happy global engine to allow various areas of the application
// to access its setup services and functions
var Bot *Engine

// New starts a new engine
func New() (*Engine, error) {
	newEngineMutex.Lock()
	defer newEngineMutex.Unlock()
	var b Engine
	b.Config = config.GetConfig()

	err := b.Config.LoadConfig("", false)
	if err != nil {
		return nil, fmt.Errorf("failed to load config. Err: %s", err)
	}

	return &b, nil
}

// NewFromSettings starts a new engine based on supplied settings
func NewFromSettings(settings *Settings, flagSet map[string]bool) (*Engine, error) {
	newEngineMutex.Lock()
	defer newEngineMutex.Unlock()
	if settings == nil {
		return nil, errors.New("engine: settings is nil")
	}

	var b Engine
	var err error

	b.Config, err = loadConfigWithSettings(settings, flagSet)
	if err != nil {
		return nil, fmt.Errorf("failed to load config. Err: %w", err)
	}

	if *b.Config.Logging.Enabled {
		err = gctlog.SetupGlobalLogger(b.Config.Name, b.Config.Logging.AdvancedSettings.StructuredLogging)
		if err != nil {
			return nil, fmt.Errorf("failed to setup global logger. %w", err)
		}
		err = gctlog.SetupSubLoggers(b.Config.Logging.SubLoggers)
		if err != nil {
			return nil, fmt.Errorf("failed to setup sub loggers. %w", err)
		}
		gctlog.Infoln(gctlog.Global, "Logger initialised.")
	}

	b.Settings.ConfigFile = settings.ConfigFile
	b.Settings.DataDir = b.Config.GetDataPath()
	b.Settings.CheckParamInteraction = settings.CheckParamInteraction

	err = utils.AdjustGoMaxProcs(settings.GoMaxProcs)
	if err != nil {
		return nil, fmt.Errorf("unable to adjust runtime GOMAXPROCS value. Err: %w", err)
	}

	validateSettings(&b, settings, flagSet)

	return &b, nil
}

// FlagSet defines set flags from command line args for comparison methods
type FlagSet map[string]bool

// WithBool checks the supplied flag. If set it will override the config boolean
// value as a command line takes precedence. If not set will fall back to config
// options.
func (f FlagSet) WithBool(key string, flagValue *bool, configValue bool) {
	isSet := f[key]
	*flagValue = !isSet && configValue || isSet && *flagValue
}

// validateSettings validates and sets all bot settings
func validateSettings(b *Engine, s *Settings, flagSet FlagSet) {
	b.Settings = *s

	// flagSet.WithBool("coinmarketcap", &b.Settings.EnableCoinmarketcapAnalysis, b.Config.Currency.CryptocurrencyProvider.Enabled)
	// flagSet.WithBool("ordermanager", &b.Settings.EnableOrderManager, b.Config.OrderManager.Enabled != nil && *b.Config.OrderManager.Enabled)

	// flagSet.WithBool("currencyconverter", &b.Settings.EnableCurrencyConverter, b.Config.Currency.ForexProviders.IsEnabled("currencyconverter"))

	// flagSet.WithBool("currencylayer", &b.Settings.EnableCurrencyLayer, b.Config.Currency.ForexProviders.IsEnabled("currencylayer"))
	// flagSet.WithBool("exchangerates", &b.Settings.EnableExchangeRates, b.Config.Currency.ForexProviders.IsEnabled("exchangerates"))
	// flagSet.WithBool("fixer", &b.Settings.EnableFixer, b.Config.Currency.ForexProviders.IsEnabled("fixer"))
	// flagSet.WithBool("openexchangerates", &b.Settings.EnableOpenExchangeRates, b.Config.Currency.ForexProviders.IsEnabled("openexchangerates"))

	// flagSet.WithBool("datahistorymanager", &b.Settings.EnableDataHistoryManager, b.Config.DataHistoryManager.Enabled)
	// flagSet.WithBool("currencystatemanager", &b.Settings.EnableCurrencyStateManager, b.Config.CurrencyStateManager.Enabled != nil && *b.Config.CurrencyStateManager.Enabled)
	// flagSet.WithBool("gctscriptmanager", &b.Settings.EnableGCTScriptManager, b.Config.GCTScript.Enabled)

	// flagSet.WithBool("tickersync", &b.Settings.EnableTickerSyncing, b.Config.SyncManagerConfig.SynchronizeTicker)
	// flagSet.WithBool("orderbooksync", &b.Settings.EnableOrderbookSyncing, b.Config.SyncManagerConfig.SynchronizeOrderbook)
	// flagSet.WithBool("tradesync", &b.Settings.EnableTradeSyncing, b.Config.SyncManagerConfig.SynchronizeTrades)
	// flagSet.WithBool("synccontinuously", &b.Settings.SyncContinuously, b.Config.SyncManagerConfig.SynchronizeContinuously)
	// flagSet.WithBool("syncmanager", &b.Settings.EnableExchangeSyncManager, b.Config.SyncManagerConfig.Enabled)

	// if b.Settings.EnablePortfolioManager &&
	// 	b.Settings.PortfolioManagerDelay <= 0 {
	// 	b.Settings.PortfolioManagerDelay = PortfolioSleepDelay
	// }

	flagSet.WithBool("grpc", &b.Settings.EnableGRPC, b.Config.RemoteControl.GRPC.Enabled)
	flagSet.WithBool("grpcproxy", &b.Settings.EnableGRPCProxy, b.Config.RemoteControl.GRPC.GRPCProxyEnabled)

	flagSet.WithBool("grpcshutdown", &b.Settings.EnableGRPCShutdown, b.Config.RemoteControl.GRPC.GRPCAllowBotShutdown)
	// if b.Settings.EnableGRPCShutdown {
	// 	b.GRPCShutdownSignal = make(chan struct{})
	// 	go b.waitForGPRCShutdown()
	// }

	flagSet.WithBool("websocketrpc", &b.Settings.EnableWebsocketRPC, b.Config.RemoteControl.WebsocketRPC.Enabled)
	flagSet.WithBool("deprecatedrpc", &b.Settings.EnableDeprecatedRPC, b.Config.RemoteControl.DeprecatedRPC.Enabled)

	// if flagSet["maxvirtualmachines"] {
	// 	maxMachines := uint8(b.Settings.MaxVirtualMachines)
	// 	b.gctScriptManager.MaxVirtualMachines = &maxMachines
	// }

	// if flagSet["withdrawcachesize"] {
	// 	withdraw.CacheSize = b.Settings.WithdrawCacheSize
	// }

	// if b.Settings.EnableEventManager && b.Settings.EventManagerDelay <= 0 {
	// 	b.Settings.EventManagerDelay = EventSleepDelay
	// }

	// if b.Settings.TradeBufferProcessingInterval != trade.DefaultProcessorIntervalTime {
	// 	if b.Settings.TradeBufferProcessingInterval >= time.Second {
	// 		trade.BufferProcessorIntervalTime = b.Settings.TradeBufferProcessingInterval
	// 	} else {
	// 		b.Settings.TradeBufferProcessingInterval = trade.DefaultProcessorIntervalTime
	// 		gctlog.Warnf(gctlog.Global, "-tradeprocessinginterval must be >= to 1 second, using default value of %v",
	// 			trade.DefaultProcessorIntervalTime)
	// 	}
	// }

	// if b.Settings.RequestMaxRetryAttempts != request.DefaultMaxRetryAttempts &&
	// 	b.Settings.RequestMaxRetryAttempts > 0 {
	// 	request.MaxRetryAttempts = b.Settings.RequestMaxRetryAttempts
	// }

	if b.Settings.HTTPTimeout <= 0 {
		b.Settings.HTTPTimeout = b.Config.GlobalHTTPTimeout
	}

	if b.Settings.GlobalHTTPTimeout <= 0 {
		b.Settings.GlobalHTTPTimeout = b.Config.GlobalHTTPTimeout
	}

	// err := common.SetHTTPClientWithTimeout(b.Settings.GlobalHTTPTimeout)
	// if err != nil {
	// 	gctlog.Errorf(gctlog.Global,
	// 		"Could not set new HTTP Client with timeout %s error: %v",
	// 		b.Settings.GlobalHTTPTimeout,
	// 		err)
	// }

	// if b.Settings.GlobalHTTPUserAgent != "" {
	// 	err = common.SetHTTPUserAgent(b.Settings.GlobalHTTPUserAgent)
	// 	if err != nil {
	// 		gctlog.Errorf(gctlog.Global, "Could not set HTTP User Agent for %s error: %v",
	// 			b.Settings.GlobalHTTPUserAgent,
	// 			err)
	// 	}
	// }

	// if b.Settings.AlertSystemPreAllocationCommsBuffer != alert.PreAllocCommsDefaultBuffer {
	// 	err = alert.SetPreAllocationCommsBuffer(b.Settings.AlertSystemPreAllocationCommsBuffer)
	// 	if err != nil {
	// 		gctlog.Errorf(gctlog.Global, "Could not set alert pre-allocation comms buffer to %v: %v",
	// 			b.Settings.AlertSystemPreAllocationCommsBuffer,
	// 			err)
	// 	}
	// }

}

// Start starts the engine
func (bot *Engine) Start() error {
	if bot == nil {
		return errors.New("engine instance is nil")
	}
	newEngineMutex.Lock()
	defer newEngineMutex.Unlock()

	gctlog.Debugln(gctlog.Global, "Engine starting...")

	// 在这里可以添加必要的初始化代码

	if bot.Settings.EnableDatabaseManager {
		if d, err := SetupDatabaseConnectionManager(&bot.Config.Database); err != nil {
			gctlog.Errorf(gctlog.Global, "Database manager unable to setup: %v", err)
		} else {
			bot.DatabaseManager = d
			if err := bot.DatabaseManager.Start(&bot.ServicesWG); err != nil && !errors.Is(err, database.ErrDatabaseSupportDisabled) {
				gctlog.Errorf(gctlog.Global, "Database manager unable to start: %v", err)
			}
		}
	}

	if bot.Settings.EnableGRPC {
		go StartRPCServer(bot)
	}

	return nil
}

// Stop correctly shuts down engine saving configuration files
func (bot *Engine) Stop() {
	newEngineMutex.Lock()
	defer newEngineMutex.Unlock()

	gctlog.Debugln(gctlog.Global, "Engine shutting down..")

	// 在这里可以添加必要的清理代码
	if bot.DatabaseManager.IsRunning() {
		if err := bot.DatabaseManager.Stop(); err != nil {
			gctlog.Errorf(gctlog.Global, "Database manager unable to stop. Error: %v", err)
		}
	}

	// Wait for services to gracefully shutdown
	bot.ServicesWG.Wait()
	gctlog.Infoln(gctlog.Global, "Exiting.")
	if err := gctlog.CloseLogger(); err != nil {
		log.Printf("Failed to close logger. Error: %v\n", err)
	}
}

// loadConfigWithSettings creates configuration based on the provided settings
func loadConfigWithSettings(settings *Settings, flagSet map[string]bool) (*config.Config, error) {
	filePath, err := config.GetAndMigrateDefaultPath(settings.ConfigFile)
	if err != nil {
		return nil, err
	}
	log.Printf("Loading config file %s..\n", filePath)

	conf := &config.Config{}
	err = conf.ReadConfigFromFile(filePath, settings.EnableDryRun)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", config.ErrFailureOpeningConfig, filePath, err)
	}
	// Apply overrides from settings
	if flagSet["datadir"] {
		// warn if dryrun isn't enabled
		if !settings.EnableDryRun {
			log.Println("Command line argument '-datadir' induces dry run mode.")
		}
		settings.EnableDryRun = true
		conf.DataDirectory = settings.DataDir
	}

	return conf, conf.CheckConfig()
}
