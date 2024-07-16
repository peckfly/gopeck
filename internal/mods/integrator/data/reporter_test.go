package data

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/integrator/biz"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

func BuildClickhouseClient() (driver.Conn, error) {
	var dialCount int
	client, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{
			Database: "gopeck",
			Username: "clickhouse",
			Password: "happy123",
		},
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			dialCount++
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Debug: true,
		Debugf: func(format string, v ...any) {
			fmt.Printf(format+"\n", v...)
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          time.Second * 30,
		MaxOpenConns:         5,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		ClientInfo: clickhouse.ClientInfo{ // optional, please see Client info section in the README.md
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "my-app", Version: "0.1"},
			},
		},
	})
	return client, err
}

func TestClickhouseQuery(t *testing.T) {
	client, err := BuildClickhouseClient()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := client.Query(context.Background(), "SELECT * FROM gopeck.test LIMIT 1")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uint32
		var name string

		if err := rows.Scan(&id, &name); err != nil {
			t.Error(err)
		}
		fmt.Printf("id: %d, name: %s\n", id, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func TestClickHouseInsert(t *testing.T) {
	client, err := BuildClickhouseClient()
	if err != nil {
		log.Fatal(err)
	}
	insertSql := "INSERT INTO gopeck.test (id, name) VALUES (?, ?)"
	err = client.Exec(context.Background(), insertSql, 2, "tx")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("insert success")
}

func TestNewReporterRepository(t *testing.T) {
	client, err := BuildClickhouseClient()
	if err != nil {
		log.Fatal(err)
	}
	repository := NewReporterRepository(client, &conf.ServerConf{
		Data: conf.DataConf{
			Clickhouse: conf.ClickhouseConf{
				StressReporterTableName: "gopeck.stress_log",
			},
		},
	})
	err = repository.Report(context.Background(), []*biz.Report{
		{
			PlanId: 1,
			TaskId: 2,
			Url:    "http://127.0.0.1",
		},
		{
			PlanId: 1,
			TaskId: 3,
			Url:    "http://127.0.0.1",
		},
	})
	assert.Equal(t, err, nil)
}
