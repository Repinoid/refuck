package models

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
type Gauge float64
type Counter int64

type MemoryStorageStruct struct {
	Gaugemetr map[string]Gauge
	Countmetr map[string]Counter
	Mutter    sync.RWMutex
}

type Inter interface {
	GetMetric(ctx context.Context, db *pgx.Conn, memorial *MemoryStorageStruct, metr *Metrics) (Metrics, error)
	PutMetric(ctx context.Context, db *pgx.Conn, memorial *MemoryStorageStruct, metr *Metrics) error
	GetAllMetrics(ctx context.Context, db *pgx.Conn, memorial *MemoryStorageStruct) (*[]Metrics, error)
	PutAllMetrics(ctx context.Context, db *pgx.Conn, memorial *MemoryStorageStruct, metras *[]Metrics) error
}
