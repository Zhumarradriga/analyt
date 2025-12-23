package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"analytservice/internal/domain"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// SQL Schema для справки:
/*
CREATE TABLE IF NOT EXISTS events (
    timestamp DateTime64(3),
    key LowCardinality(String),
    properties String
) ENGINE = MergeTree()
ORDER BY (key, timestamp)
PARTITION BY toYYYYMM(timestamp);
*/

type EventRepo interface {
	SaveBatch(ctx context.Context, events []domain.Event) error
	Close() error
}

type clickHouseRepo struct {
	conn driver.Conn
}

func NewClickHouseRepo(addr, db, user, password string) (EventRepo, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: db,
			Username: user,
			Password: password,
		},
		Compression: &clickhouse.Compression{Method: clickhouse.CompressionLZ4},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	})

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &clickHouseRepo{conn: conn}, nil
}

func (r *clickHouseRepo) SaveBatch(ctx context.Context, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO events (timestamp, key, properties)")
	if err != nil {
		return fmt.Errorf("prepare batch error: %w", err)
	}

	for _, e := range events {
		propsBytes, _ := json.Marshal(e.Value)
		
		err := batch.Append(
			e.Timestamp,
			e.Key,
			string(propsBytes),
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func (r *clickHouseRepo) Close() error {
	return r.conn.Close()
}