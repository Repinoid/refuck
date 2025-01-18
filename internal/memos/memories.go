package memos

import (
	"context"
	"fmt"
	"gorono/internal/models"
	"log"

	"github.com/jackc/pgx/v5"
)

type Metrics = models.Metrics
type MemStruct struct {
}

func (gag MemStruct) PutMetric(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct, metr *Metrics) error {
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()
	switch metr.MType {
	case "gauge":
		memorial.Gaugemetr[metr.ID] = models.Gauge(*metr.Value)
	case "counter":
		memorial.Countmetr[metr.ID] = models.Counter(*metr.Delta)
	default:
		return fmt.Errorf("wrong metric %+v", metr)
	}
	return nil
}

func (gag MemStruct) GetMetric(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct, metr *Metrics) (Metrics, error) {
	memorial.Mutter.RLock() // <---- MUTEX
	defer memorial.Mutter.RUnlock()
	switch metr.MType {
	case "gauge":
		if val, ok := memorial.Gaugemetr[metr.ID]; ok {
			out := float64(val)
			metr.Value = &out
		}
	case "counter":
		if val, ok := memorial.Countmetr[metr.ID]; ok {
			out := int64(val)
			metr.Delta = &out
		}
	default:
		return *metr, fmt.Errorf("wrong metric %+v", metr)
	}
	return *metr, nil
}

// --- from []Metrics to memory Storage
func (gag MemStruct) PutAllMetrics(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct, metras *[]Metrics) error {
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()

	for _, metr := range *metras {
		switch metr.MType {
		case "gauge":
			memorial.Gaugemetr[metr.ID] = gauge(*metr.Value)
		case "counter":
			if _, ok := memorial.Countmetr[metr.ID]; ok {
				memorial.Countmetr[metr.ID] += counter(*metr.Delta)
				continue
			}
			memorial.Countmetr[metr.ID] = counter(*metr.Delta)
		default:
			log.Printf("wrong metric type %s\n", metr.MType)
		}
	}
	return nil
}

// ----- from Memory Storage to []Metrics
func (gag MemStruct) GetAllMetrics(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct) (*[]Metrics, error) {

	memorial.Mutter.RLock()
	defer memorial.Mutter.RUnlock()

	metras := []Metrics{}

	for nam, val := range memorial.Countmetr {
		out := int64(val)
		metr := Metrics{ID: nam, MType: "counter", Delta: &out}
		metras = append(metras, metr)
	}
	for nam, val := range memorial.Gaugemetr {
		out := float64(val)
		metr := Metrics{ID: nam, MType: "counter", Value: &out}
		metras = append(metras, metr)
	}
	return &metras, nil

}
