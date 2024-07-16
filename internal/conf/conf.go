package conf

import (
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/spf13/viper"
	"os"
)

type (
	ServerConf struct {
		Name       string           `mapstructure:"name"`
		Rbac       RbacConf         `mapstructure:"rbac"`
		Casbin     CasbinConf       `mapstructure:"casbin"`
		Auth       AuthConf         `mapstructure:"auth"`
		Log        log.LoggerConfig `mapstructure:"log"`
		Server     Server           `mapstructure:"server"`
		StressConf WorkerStressConf `mapstructure:"stress_config"`
		Data       DataConf         `mapstructure:"data"`
		Metrics    MetricsConf      `mapstructure:"metrics"`
	}
	RbacConf struct {
		RootUsername    string `mapstructure:"root_username"`
		RootName        string `mapstructure:"root_name"`
		RootPassword    string `mapstructure:"root_password"`
		RootId          string `mapstructure:"root_id"`
		UserCacheExp    int64  `mapstructure:"user_cache_exp"`
		DenyDeleteMenu  bool   `mapstructure:"deny_delete_menu"`
		DefaultLoginPwd string `mapstructure:"default_login_pwd"`
		AuthDisable     bool   `mapstructure:"auth_disable"`
	}

	AuthConf struct {
		Disable             bool     `mapstructure:"disable"`
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		SigningMethod       string   `mapstructure:"signing_method"`  // HS256/HS384/HS512
		SigningKey          string   `mapstructure:"signing_key"`     // secret key
		OldSigningKey       string   `mapstructure:"old_signing_key"` // old secret key (for migration)
		Expired             int      `mapstructure:"expired"`         // seconds
	}

	CasbinConf struct {
		Disable             bool     `mapstructure:"disable"`
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		LoadThread          int      `mapstructure:"load_thread"`
		AutoLoadInterval    int      `mapstructure:"auto_load_interval"`
		ModelFile           string   `mapstructure:"model_file"`
		GenPolicyFile       string   `mapstructure:"gen_policy_file"`
		WorkDir             string   `mapstructure:"work_dir"`
	}

	WorkerStressConf struct {
		// pecker max goroutine num every node, testing number before deployment
		MaxGoroutineNum int `mapstructure:"max_goroutine_num"`
		// pecker max rps every node, testing number before deployment
		MaxRpsNum             int   `mapstructure:"max_rps_num"`
		MaxResultChanSize     int32 `mapstructure:"max_result_chan_size"`
		RpsResultChanBlowup   int32 `mapstructure:"rps_result_chan_blowup"`
		MaxTimeoutSecond      int64 `mapstructure:"max_timeout_second"`
		ReportGoroutineNum    int   `mapstructure:"report_goroutine_num"`
		DefaultMaxConnections int   `mapstructure:"default_max_connections"`
		ErrorCutLength        int   `mapstructure:"error_cut_length"`
	}

	Server struct {
		Http HttpConf `mapstructure:"http"`
		Grpc GrpcConf `mapstructure:"grpc"`
	}

	HttpConf struct {
		Port       int    `mapstructure:"port"`
		JwtSignKey string `mapstructure:"jwt_sign_key"`
	}

	GrpcConf struct {
		Addr string `mapstructure:"addr"`
	}

	DataConf struct {
		Database   DatabaseConf   `mapstructure:"database"`
		Redis      RedisConf      `mapstructure:"redis"`
		Etcd       EtcdConf       `mapstructure:"etcd"`
		Clickhouse ClickhouseConf `mapstructure:"clickhouse"`
	}

	DatabaseConf struct {
		Dsn          string `mapstructure:"dsn"`
		MaxIdleConns int    `mapstructure:"max_idle_conns"`
		MaxOpenConns int    `mapstructure:"max_open_conns"`
	}

	RedisConf struct {
		Type           int    `mapstructure:"type"`
		Addr           string `mapstructure:"addr"`
		Password       string `mapstructure:"password"`
		DB             int    `mapstructure:"db"`
		DialTimeout    int    `mapstructure:"dial_timeout"`
		ReadTimeout    int    `mapstructure:"read_timeout"`
		WriteTimeout   int    `mapstructure:"write_timeout"`
		MinIdleConns   int    `mapstructure:"min_idle_conns"`
		MaxIdleConns   int    `mapstructure:"max_idle_conns"`
		MaxActiveConns int    `mapstructure:"max_active_conns"`
	}

	EtcdConf struct {
		Endpoints string `mapstructure:"endpoints"`
		Username  string `mapstructure:"username"`
		Password  string `mapstructure:"password"`
	}

	ClickhouseConf struct {
		Addrs                   string `mapstructure:"addrs"`
		Database                string `mapstructure:"database"`
		Username                string `mapstructure:"username"`
		Password                string `mapstructure:"password"`
		MaxOpenConns            int    `mapstructure:"max_open_conns"`
		MaxIdleConns            int    `mapstructure:"max_idle_conns"`
		StressReporterTableName string `mapstructure:"stress_reporter_table_name"`
	}

	MetricsConf struct {
		GrafanaAddr string `mapstructure:"grafana_addr"`
	}
)

// ReadConfig read config from path
func ReadConfig(path string) *ServerConf {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	var conf ServerConf
	if err := v.Unmarshal(&conf); err != nil {
		panic(err)
	}
	// handle environment variables
	conf.Data.Database.Dsn = os.ExpandEnv(conf.Data.Database.Dsn)
	conf.Data.Redis.Addr = os.ExpandEnv(conf.Data.Redis.Addr)
	conf.Data.Clickhouse.Addrs = os.ExpandEnv(conf.Data.Clickhouse.Addrs)
	conf.Data.Etcd.Endpoints = os.ExpandEnv(conf.Data.Etcd.Endpoints)
	return &conf
}
