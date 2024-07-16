package initialize

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-jwt/jwt"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/internal/pkg/jwtx"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	logsdk "log"
	"net"
	"strings"
	"time"
)

// NewCkClient create clickhouse client with data config
func NewCkClient(conf *conf.ServerConf) driver.Conn {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: strings.Split(conf.Data.Clickhouse.Addrs, ","),
		Auth: clickhouse.Auth{
			Database: conf.Data.Clickhouse.Database,
			Username: conf.Data.Clickhouse.Username,
			Password: conf.Data.Clickhouse.Password,
		},
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Debug: false,
		Debugf: func(format string, v ...any) {
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          time.Second * 30,
		MaxOpenConns:         conf.Data.Clickhouse.MaxOpenConns,
		MaxIdleConns:         conf.Data.Clickhouse.MaxIdleConns,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "gopeck", Version: "0.1"},
			},
		},
	})
	log.Must(err)
	return conn
}

// NewEtcdClient create etcd client with data config
func NewEtcdClient(conf *conf.ServerConf) *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(conf.Data.Etcd.Endpoints, ","),
		Username:    conf.Data.Etcd.Username,
		Password:    conf.Data.Etcd.Password,
		DialTimeout: 5 * time.Second,
	})
	log.Must(err)
	return client
}

// NewRedisCacheClient create redis client with data config
func NewRedisCacheClient(conf *conf.ServerConf) cachex.Cache {
	if conf.Data.Redis.Type == enums.RedisSingle {
		return cachex.NewRedisCacheWithClient(redis.NewClient(&redis.Options{
			Addr:           conf.Data.Redis.Addr,
			Password:       conf.Data.Redis.Password,
			DB:             conf.Data.Redis.DB,
			DialTimeout:    time.Duration(conf.Data.Redis.DialTimeout) * time.Millisecond,
			ReadTimeout:    time.Duration(conf.Data.Redis.ReadTimeout) * time.Millisecond,
			WriteTimeout:   time.Duration(conf.Data.Redis.WriteTimeout) * time.Millisecond,
			MinIdleConns:   conf.Data.Redis.MinIdleConns,
			MaxIdleConns:   conf.Data.Redis.MaxIdleConns,
			MaxActiveConns: conf.Data.Redis.MaxActiveConns,
		}))
	} else {
		addrs := strings.Split(conf.Data.Redis.Addr, ",")
		return cachex.NewRedisCacheWithClusterClient(redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:          addrs,
			Password:       conf.Data.Redis.Password,
			DialTimeout:    time.Duration(conf.Data.Redis.DialTimeout) * time.Millisecond,
			ReadTimeout:    time.Duration(conf.Data.Redis.ReadTimeout) * time.Millisecond,
			WriteTimeout:   time.Duration(conf.Data.Redis.WriteTimeout) * time.Millisecond,
			MinIdleConns:   conf.Data.Redis.MinIdleConns,
			MaxIdleConns:   conf.Data.Redis.MaxIdleConns,
			MaxActiveConns: conf.Data.Redis.MaxActiveConns,
		}))
	}
}

// NewDbClient create db client with data config
func NewDbClient(conf *conf.ServerConf) *gorm.DB {
	dsn := conf.Data.Database.Dsn
	mysqlConfig := mysql.Config{
		DSN: dsn,
	}
	mysqlConfig.SkipInitializeWithVersion = true
	dbLogger := logger.New(
		logsdk.New(log.NewZapInfoWriter(), "", logsdk.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second * 3,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)
	gormConfig := &gorm.Config{
		Logger: dbLogger,
	}
	db, err := gorm.Open(mysql.New(mysqlConfig), gormConfig)
	sqlDb, err := db.DB()
	sqlDb.SetMaxIdleConns(conf.Data.Database.MaxIdleConns)
	sqlDb.SetMaxOpenConns(conf.Data.Database.MaxOpenConns)
	log.Must(err)
	return db
}

func InitAuth(ctx context.Context, cache cachex.Cache, authConf *conf.AuthConf) (jwtx.Auther, func(), error) {
	var opts []jwtx.Option
	opts = append(opts, jwtx.SetExpired(authConf.Expired))
	opts = append(opts, jwtx.SetSigningKey(authConf.SigningKey, authConf.OldSigningKey))

	var method jwt.SigningMethod
	switch authConf.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	default:
		method = jwt.SigningMethodHS512
	}
	opts = append(opts, jwtx.SetSigningMethod(method))

	auth := jwtx.New(jwtx.NewStoreWithCache(cache), opts...)
	return auth, func() {
		_ = auth.Release(ctx)
	}, nil
}
