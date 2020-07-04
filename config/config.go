package config

import (
	"github.com/siddontang/go-log/log"
	"github.com/spf13/viper"
)

type cfgInfo struct {
	LogConfig       *LogConfigInfo     `yaml:"logconfig"`
	AddressConfig   *AddressConfigInfo `yaml:"addressconfig"`
	P2PConfig       *P2PConfigInfo     `yaml:"p2pconfig"`
	ConsensusConfig *BftConfig         `yaml:"consensusconfig"`
	APIConfig       *APIConfigInfo     `yaml:"apiconfig"`
}

type LogConfigInfo struct {
	Level      string `yaml:"level"`
	FileName   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`
	MaxAge     int    `yaml:"maxage"`
	MaxBackups int    `yaml:"maxbackups"`
	Comperss   bool   `yaml:"comperss"`
}
type RPCConfigInfo struct {
	Address  string `yaml:"address"`
	CertFile string `yaml:"certfile"`
	KeyFile  string `yaml:"keyfile"`
}

type WEBConfigInfo struct {
	Address string `yaml:"address"`
}

type APIConfigInfo struct {
	Port      string         `yaml:"port"`
	RPCConfig *RPCConfigInfo `yaml:"rpcconfig"`
	WEBConfig *WEBConfigInfo `yaml:"webconfig"`
}

type P2PConfigInfo struct {
	BindPort      int      `yaml:"bindport"`
	BindAddr      string   `yaml:"bindaddr"`
	AdvertiseAddr string   `yaml:"advertiseaddr"`
	NodeName      string   `yaml:"nodename"`
	Members       []string `yaml:"members"`
}

type AddressConfigInfo struct {
	QTJAddress   string `yaml:"qtjaddress"`
	DSAddress    string `yaml:"dsaddress"`
	CMAddress    string `yaml:"cmaddress"`
	MinerAddress string `yamil:"mineraddress"`
}

type ConsensusConfigInfo struct {
	Id        int      `yaml:"id"`
	Address   string   `yaml:"address"`
	Peer      string   `yaml:"peer"`
	Peers     []string `yaml:"peers"`
	Join      bool     `yaml:"join"`
	Waldir    string   `yaml:"waldir"`
	Snapdir   string   `yaml:"snapdir"`
	Raftport  int64    `yaml:"raftport"`
	Ds        string   `yaml:"ds"`
	Cm        string   `yaml:"cm"`
	QTJ       string   `yaml:"qtj"`
	SnapCount int64    `yaml:"snapcount"`

	LogFile     string `yaml:"logfile"`
	LogSaveDays int    `yaml:"logsavedays"`
	LogLevel    int    `yaml:"loglevel"`
	LogSaveMode int    `yaml:"logsavemode"`
	LogFileSize int64  `yaml:"logfilesize"`
}

type BftConfig struct {
	NodeNum          uint64   `json:"nodenum"`
	Peers            []string `json:"peers"`
	HttpAddr         string   `json:"httpaddr"`
	NodeAddr         string   `json:"nodeaddr"`
	CountAddr        string   `json:"countaddr"`
	RpcAddr          string   `json:"rpcaddr"`
	Join             bool     `json:"join"`
	SnapshotCount    uint64   `json:"snapshotcount"`
	SnapshotInterval uint64   `json:"Snapshotinterval"`
	Ds               string   `json:"ds"`
	Cm               string   `json:"cm"`
	QTJ              string   `json:"qtg"`

	LogDir    string `json:"logdir"`
	SnapDir   string `json:"snapdir"`
	LogsDir   string `json:"logsdir"`
	StableDir string `json:"stabledir"`

	LogFile     string `json:"logfile"`
	LogSaveDays int    `json:"logsavedays"`
	LogLevel    int    `json:"loglevel"`
	LogSaveMode int    `json:"logsavemode"`
	LogFileSize int64  `json:"logfilesize"`
}

func LoadConfig() (*cfgInfo, error) {
	viper.SetConfigName("kortho")
	viper.AddConfigPath("./configs/")
	if err := viper.ReadInConfig(); err != nil {
		log.Panic(err)
	}

	var cfg cfgInfo
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
