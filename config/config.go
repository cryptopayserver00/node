package config

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

type Server struct {
	System           System           `mapstructure:"system" json:"system" yaml:"system"`
	Redis            Redis            `mapstructure:"redis" json:"redis" yaml:"redis"`
	Memcache         Memcache         `mapstructure:"memcache" json:"memcache" yaml:"memcache"`
	Mysql            Mysql            `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Zap              Zap              `mapstructure:"zap" json:"zap" yaml:"zap"`
	Blockchain       Blockchain       `mapstructure:"blockchain" json:"blockchain" yaml:"blockchain"`
	BlockchainPlugin BlockchainPlugin `mapstructure:"blockchain-plugin" json:"blockchain-plugin" yaml:"blockchain-plugin"`
	Telegram         Telegram         `mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Wss              Wss              `mapstructure:"wss" json:"wss" yaml:"wss"`
	FreeCoin         FreeCoin         `mapstructure:"freecoin" json:"freecoin" yaml:"freecoin"`
	Key              Key              `mapstructure:"key" json:"key" yaml:"key"`
}

type Key struct {
	TrongridMainnetKey       string `mapstructure:"trongrid_mainnet_key" json:"trongrid_mainnet_key" yaml:"trongrid_mainnet_key"`
	TrongridNileKey          string `mapstructure:"trongrid_nile_key" json:"trongrid_nile_key" yaml:"trongrid_nile_key"`
	AlchemyMainnetKey        string `mapstructure:"alchemy_mainnet_key" json:"alchemy_mainnet_key" yaml:"alchemy_mainnet_key"`
	AlchemyInnerTxMainnetKey string `mapstructure:"alchemy_inner_tx_mainnet_key" json:"alchemy_inner_tx_mainnet_key" yaml:"alchemy_inner_tx_mainnet_key"`
	AlchemyInnerTxTestnetKey string `mapstructure:"alchemy_inner_tx_testnet_key" json:"alchemy_inner_tx_testnet_key" yaml:"alchemy_inner_tx_testnet_key"`
	AlchemyTestnetKey        string `mapstructure:"alchemy_testnet_key" json:"alchemy_testnet_key" yaml:"alchemy_testnet_key"`
	TatumMainnetKey          string `mapstructure:"tatum_mainnet_key" json:"tatum_mainnet_key" yaml:"tatum_mainnet_key"`
	TatumTestnetKey          string `mapstructure:"tatum_testnet_key" json:"tatum_testnet_key" yaml:"tatum_testnet_key"`
}

type FreeCoin struct {
	Bitcoin     Bitcoin     `mapstructure:"bitcoin" json:"bitcoin" yaml:"bitcoin"`
	Ethereum    Ethereum    `mapstructure:"ethereum" json:"ethereum" yaml:"ethereum"`
	Solana      Solana      `mapstructure:"solana" json:"solana" yaml:"solana"`
	Litecoin    Litecoin    `mapstructure:"litecoin" json:"litecoin" yaml:"litecoin"`
	Tron        Tron        `mapstructure:"tron" json:"tron" yaml:"tron"`
	Ton         Ton         `mapstructure:"ton" json:"ton" yaml:"ton"`
	Xrp         Xrp         `mapstructure:"xrp" json:"xrp" yaml:"xrp"`
	BitcoinCash BitcoinCash `mapstructure:"bitcoincash" json:"bitcoincash" yaml:"bitcoincash"`
}

type Bitcoin struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	Mnemonic   string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type Ethereum struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
}

type Solana struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	Mnemonic   string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type Litecoin struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	Mnemonic   string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type Tron struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	Mnemonic   string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type Ton struct {
	PublicKey string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	Mnemonic  string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type Xrp struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	Mnemonic   string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type BitcoinCash struct {
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	Mnemonic   string `mapstructure:"mnemonic" json:"mnemonic" yaml:"mnemonic"`
}

type Telegram struct {
	InformBotLink     string `mapstructure:"inform-bot-link" json:"inform-bot-link" yaml:"inform-bot-link"`
	InformBotToken    string `mapstructure:"inform-bot-token" json:"inform-bot-token" yaml:"inform-bot-token"`
	InformChannelId   int64  `mapstructure:"inform-channel-id" json:"inform-channel-id" yaml:"inform-channel-id"`
	TxInformChannelId int64  `mapstructure:"tx-inform-channel-id" json:"tx-inform-channel-id" yaml:"tx-inform-channel-id"`
}

type Wss struct {
	SecWssToken string `mapstructure:"sec-wss-token" json:"sec-wss-token" yaml:"sec-wss-token"`
}

type Blockchain struct {
	OpenSweepBlock bool `mapstructure:"open-sweep-block" json:"open-sweep-block" yaml:"open-sweep-block"`
	SweepMainnet   bool `mapstructure:"sweep-mainnet" json:"sweep-mainnet" yaml:"sweep-mainnet"`
	Ethereum       bool `mapstructure:"ethereum" json:"ethereum" yaml:"ethereum"`
	Bitcoin        bool `mapstructure:"bitcoin" json:"bitcoin" yaml:"bitcoin"`
	Tron           bool `mapstructure:"tron" json:"tron" yaml:"tron"`
	Bsc            bool `mapstructure:"bsc" json:"bsc" yaml:"bsc"`
	Litecoin       bool `mapstructure:"litecoin" json:"litecoin" yaml:"litecoin"`
	Op             bool `mapstructure:"op" json:"op" yaml:"op"`
	ArbitrumOne    bool `mapstructure:"arbitrum-one" json:"arbitrum-one" yaml:"arbitrum-one"`
	ArbitrumNova   bool `mapstructure:"arbitrum-nova" json:"arbitrum-nova" yaml:"arbitrum-nova"`
	Solana         bool `mapstructure:"solana" json:"solana" yaml:"solana"`
	Ton            bool `mapstructure:"ton" json:"ton" yaml:"ton"`
	Xrp            bool `mapstructure:"xrp" json:"xrp" yaml:"xrp"`
	Bch            bool `mapstructure:"bch" json:"bch" yaml:"bch"`
	Pol            bool `mapstructure:"pol" json:"pol" yaml:"pol"`
	Avax           bool `mapstructure:"avax" json:"avax" yaml:"avax"`
	Base           bool `mapstructure:"base" json:"base" yaml:"base"`
}

type BlockchainPlugin struct {
	Bitcoin  string `mapstructure:"bitcoin" json:"bitcoin" yaml:"bitcoin"`
	Litecoin string `mapstructure:"litecoin" json:"litecoin" yaml:"litecoin"`
}

type System struct {
	Env          string `mapstructure:"env" json:"env" yaml:"env"`
	DbType       string `mapstructure:"db-type" json:"db-type" yaml:"db-type"`
	RouterPrefix string `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`
	Addr         int    `mapstructure:"addr" json:"addr" yaml:"addr"`
	LimitCountIP int    `mapstructure:"iplimit-count" json:"iplimit-count" yaml:"iplimit-count"`
	LimitTimeIP  int    `mapstructure:"iplimit-time" json:"iplimit-time" yaml:"iplimit-time"`
	UseRedis     bool   `mapstructure:"use-redis" json:"use-redis" yaml:"use-redis"`
	UseMemcache  bool   `mapstructure:"use-memcache" json:"use-memcache" yaml:"use-memcache"`
	UseMongo     bool   `mapstructure:"use-mongo" json:"use-mongo" yaml:"use-mongo"`
	UseInit      bool   `mapstructure:"use-init" json:"use-init" yaml:"use-init"`
	UseTask      bool   `mapstructure:"use-task" json:"use-task" yaml:"use-task"`
}

type Redis struct {
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DB       int    `mapstructure:"db" json:"db" yaml:"db"`
}

type Memcache struct {
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}

type GeneralDB struct {
	Prefix       string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	Config       string `mapstructure:"config" json:"config" yaml:"config"`
	Dbname       string `mapstructure:"db-name" json:"db-name" yaml:"db-name"`
	Username     string `mapstructure:"username" json:"username" yaml:"username"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	Path         string `mapstructure:"path" json:"path" yaml:"path"`
	Engine       string `mapstructure:"engine" json:"engine" yaml:"engine" default:"InnoDB"`
	LogMode      string `mapstructure:"log-mode" json:"log-mode" yaml:"log-mode"`
	MaxIdleConns int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"`
	Singular     bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	LogZap       bool   `mapstructure:"log-zap" json:"log-zap" yaml:"log-zap"`
}

type Mysql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

func (m *Mysql) GetLogMode() string {
	return m.LogMode
}

type Zap struct {
	Level         string `mapstructure:"level" json:"level" yaml:"level"`
	Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	Format        string `mapstructure:"format" json:"format" yaml:"format"`
	Director      string `mapstructure:"director" json:"director"  yaml:"director"`
	EncodeLevel   string `mapstructure:"encode-level" json:"encode-level" yaml:"encode-level"`
	StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktrace-key" yaml:"stacktrace-key"`
	MaxAge        int    `mapstructure:"max-age" json:"max-age" yaml:"max-age"`
	ShowLine      bool   `mapstructure:"show-line" json:"show-line" yaml:"show-line"`
	LogInConsole  bool   `mapstructure:"log-in-console" json:"log-in-console" yaml:"log-in-console"`
}

func (z *Zap) ZapEncodeLevel() zapcore.LevelEncoder {
	switch {
	case z.EncodeLevel == "LowercaseLevelEncoder":
		return zapcore.LowercaseLevelEncoder
	case z.EncodeLevel == "LowercaseColorLevelEncoder":
		return zapcore.LowercaseColorLevelEncoder
	case z.EncodeLevel == "CapitalLevelEncoder":
		return zapcore.CapitalLevelEncoder
	case z.EncodeLevel == "CapitalColorLevelEncoder":
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

func (z *Zap) TransportLevel() zapcore.Level {
	z.Level = strings.ToLower(z.Level)
	switch z.Level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}
