package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"gocryptotrader/common"
	"gocryptotrader/common/convert"
	"gocryptotrader/common/file"
	"gocryptotrader/config/versions"
	"gocryptotrader/log"
)

var (
	errExchangeConfigIsNil  = errors.New("exchange config is nil")
	errPairsManagerIsNil    = errors.New("currency pairs manager is nil")
	errDecryptFailed        = errors.New("failed to decrypt config after 3 attempts")
	errCheckingConfigValues = errors.New("fatal error checking config values")
)

// CheckLoggerConfig checks to see logger values are present and valid in config
// if not creates a default instance of the logger
func (c *Config) CheckLoggerConfig() error {
	m.Lock()
	defer m.Unlock()

	if c.Logging.Enabled == nil || c.Logging.Output == "" {
		c.Logging = *log.GenDefaultSettings()
	}

	if c.Logging.AdvancedSettings.ShowLogSystemName == nil {
		c.Logging.AdvancedSettings.ShowLogSystemName = convert.BoolPtr(false)
	}

	if c.Logging.LoggerFileConfig != nil {
		if c.Logging.LoggerFileConfig.FileName == "" {
			c.Logging.LoggerFileConfig.FileName = "log.txt"
		}
		if c.Logging.LoggerFileConfig.Rotate == nil {
			c.Logging.LoggerFileConfig.Rotate = convert.BoolPtr(false)
		}
		if c.Logging.LoggerFileConfig.MaxSize <= 0 {
			log.Warnf(log.ConfigMgr, "Logger rotation size invalid, defaulting to %v", log.DefaultMaxFileSize)
			c.Logging.LoggerFileConfig.MaxSize = log.DefaultMaxFileSize
		}
		log.SetFileLoggingState( /*Is correctly configured*/ true)
	}

	err := log.SetGlobalLogConfig(&c.Logging)
	if err != nil {
		return err
	}

	logPath := c.GetDataPath("logs")
	err = common.CreateDir(logPath)
	if err != nil {
		return err
	}
	return log.SetLogPath(logPath)
}

// DefaultFilePath returns the default config file path
// MacOS/Linux: $HOME/.gocryptotrader/config.json or config.dat
// Windows: %APPDATA%\GoCryptoTrader\config.json or config.dat
// Helpful for printing application usage
func DefaultFilePath() string {
	foundConfig, _, err := GetFilePath("")
	if err != nil {
		// If there was no config file, show default location for .json
		return filepath.Join(common.GetDefaultDataDir(runtime.GOOS), File)
	}
	return foundConfig
}

// GetFilePath returns the desired config file or the default config file name
// and whether it was loaded from a default location (rather than explicitly specified)
func GetFilePath(configFile string) (configPath string, isImplicitDefaultPath bool, err error) {
	if configFile != "" {
		return configFile, false, nil
	}

	exePath, err := common.GetExecutablePath()
	if err != nil {
		return "", false, err
	}
	newDir := common.GetDefaultDataDir(runtime.GOOS)
	defaultPaths := []string{
		filepath.Join(exePath, File),
		filepath.Join(exePath, EncryptedFile),
		filepath.Join(newDir, File),
		filepath.Join(newDir, EncryptedFile),
	}

	for _, p := range defaultPaths {
		if file.Exists(p) {
			configFile = p
			break
		}
	}
	if configFile == "" {
		return "", false, fmt.Errorf("config.json file not found in %s, please follow README.md in root dir for config generation",
			newDir)
	}

	return configFile, true, nil
}

// SaveConfigToFile saves your configuration to your desired path as a JSON object.
// The function encrypts the data and prompts for encryption key, if necessary
func (c *Config) SaveConfigToFile(configPath string) error {
	defaultPath, _, err := GetFilePath(configPath)
	if err != nil {
		return err
	}
	var writer *os.File
	provider := func() (io.Writer, error) {
		writer, err = file.Writer(defaultPath)
		return writer, err
	}
	defer func() {
		if writer != nil {
			err = writer.Close()
			if err != nil {
				log.Errorln(log.ConfigMgr, err)
			}
		}
	}()
	return c.Save(provider)
}

// Save saves your configuration to the writer as a JSON object with encryption, if configured
// If there is an error when preparing the data to store, the writer is never requested
func (c *Config) Save(writerProvider func() (io.Writer, error)) error {
	payload, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}

	if c.EncryptConfig == fileEncryptionEnabled {
		// Ensure we have the key from session or from user
		if len(c.sessionDK) == 0 {
			f := c.EncryptionKeyProvider
			if f == nil {
				f = PromptForConfigKey
			}
			var key, sessionDK, storedSalt []byte
			if key, err = f(true); err != nil {
				return err
			}
			if sessionDK, storedSalt, err = makeNewSessionDK(key); err != nil {
				return err
			}
			c.sessionDK, c.storedSalt = sessionDK, storedSalt
		}
		payload, err = c.encryptConfigFile(payload)
		if err != nil {
			return err
		}
	}
	configWriter, err := writerProvider()
	if err != nil {
		return err
	}
	_, err = io.Copy(configWriter, bytes.NewReader(payload))
	return err
}

// CheckConfig checks all config settings
func (c *Config) CheckConfig() error {
	if err := c.CheckLoggerConfig(); err != nil {
		log.Errorf(log.ConfigMgr, "Failed to configure logger, some logging features unavailable: %s\n", err)
	}

	if c.GlobalHTTPTimeout <= 0 {
		log.Warnf(log.ConfigMgr, "Global HTTP Timeout value not set, defaulting to %v.\n", defaultHTTPTimeout)
		c.GlobalHTTPTimeout = defaultHTTPTimeout
	}

	return nil
}

// LoadConfig loads your configuration file into your configuration object
func (c *Config) LoadConfig(configPath string, dryrun bool) error {
	err := c.ReadConfigFromFile(configPath, dryrun)
	if err != nil {
		return fmt.Errorf("%w (%s): %w", ErrFailureOpeningConfig, configPath, err)
	}
	return c.CheckConfig()
}

// ReadConfigFromFile loads Config from the path
// If encrypted, prompts for encryption key
// Unless dryrun checks if the configuration needs to be encrypted and resaves, prompting for key
func (c *Config) ReadConfigFromFile(path string, dryrun bool) error {
	var err error
	path, _, err = GetFilePath(path)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := c.readConfig(f); err != nil {
		return err
	}

	if dryrun || c.EncryptConfig != fileEncryptionPrompt || IsFileEncrypted(path) {
		return nil
	}

	return c.saveWithEncryptPrompt(path)
}

// saveWithEncryptPrompt will prompt the user if they want to encrypt their config
// If they agree, c.EncryptConfig is set to Enabled, the config is encrypted and saved
// Otherwise, c.EncryptConfig is set to Disabled and the file is resaved
func (c *Config) saveWithEncryptPrompt(path string) error {
	if confirm, err := promptForConfigEncryption(); err != nil {
		return nil //nolint:nilerr // Ignore encryption prompt failures; The user will be prompted again
	} else if confirm {
		c.EncryptConfig = fileEncryptionEnabled
		return c.SaveConfigToFile(path)
	}

	c.EncryptConfig = fileEncryptionDisabled
	return c.SaveConfigToFile(path)
}

// readConfig loads config from a io.Reader into the config object
// versions manager will upgrade/downgrade if appropriate
// If encrypted, prompts for encryption key
func (c *Config) readConfig(d io.Reader) error {
	j, err := io.ReadAll(d)
	if err != nil {
		return err
	}

	if j, err = versions.Manager.Deploy(context.Background(), j); err != nil {
		return err
	}

	return json.Unmarshal(j, c)
}

// UpdateConfig updates the config with a supplied config file
func (c *Config) UpdateConfig(configPath string, newCfg *Config, dryrun bool) error {
	err := newCfg.CheckConfig()
	if err != nil {
		return err
	}

	c.Name = newCfg.Name
	c.EncryptConfig = newCfg.EncryptConfig
	c.GlobalHTTPTimeout = newCfg.GlobalHTTPTimeout

	if !dryrun {
		err = c.SaveConfigToFile(configPath)
		if err != nil {
			return err
		}
	}

	return c.LoadConfig(configPath, dryrun)
}

// GetConfig returns the global shared config instance
func GetConfig() *Config {
	m.Lock()
	defer m.Unlock()
	return &cfg
}

// SetConfig sets the global shared config instance
func SetConfig(c *Config) {
	m.Lock()
	defer m.Unlock()
	cfg = *c
}

// GetDataPath gets the data path for the given subpath
func (c *Config) GetDataPath(elem ...string) string {
	var baseDir string
	if c.DataDirectory != "" {
		baseDir = c.DataDirectory
	} else {
		baseDir = common.GetDefaultDataDir(runtime.GOOS)
	}
	return filepath.Join(append([]string{baseDir}, elem...)...)
}

// Validate checks if exchange config is valid
func (c *Exchange) Validate() error {
	if c == nil {
		return errExchangeConfigIsNil
	}
	return nil
}

// GetAndMigrateDefaultPath returns the target config file
// migrating it from the old default location to new one,
// if it was implicitly loaded from a default location and
// wasn't already in the correct 'new' default location
func GetAndMigrateDefaultPath(configFile string) (string, error) {
	filePath, wasDefault, err := GetFilePath(configFile)
	if err != nil {
		return "", err
	}
	if wasDefault {
		return migrateConfig(filePath, common.GetDefaultDataDir(runtime.GOOS))
	}
	return filePath, nil
}

// migrateConfig will move the config file to the target
// config directory as `File` or `EncryptedFile` depending on whether the config
// is encrypted
func migrateConfig(configFile, targetDir string) (string, error) {
	var target string
	if IsFileEncrypted(configFile) {
		target = EncryptedFile
	} else {
		target = File
	}
	target = filepath.Join(targetDir, target)
	if configFile == target {
		return configFile, nil
	}
	if file.Exists(target) {
		log.Warnf(log.ConfigMgr, "config file already found in '%s'; not overwriting, defaulting to %s", target, configFile)
		return configFile, nil
	}

	if err := file.Move(configFile, target); err != nil {
		return "", err
	}

	return target, nil
}
