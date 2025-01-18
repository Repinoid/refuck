package memos

import (
	"context"
	"fmt"
	"gorono/internal/models"
	"log"
	"sync"
)

type MemoryStorageStruct struct {
	Gaugemetr map[string]models.Gauge
	Countmetr map[string]models.Counter
	mutter    *sync.RWMutex
}
type Metrics = models.Metrics

func (memorial MemoryStorageStruct) PutMetric(ctx context.Context, metr *Metrics) error {
	memorial.mutter.Lock()
	defer memorial.mutter.Unlock()
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

func (memorial MemoryStorageStruct) GetMetric(ctx context.Context, metr *Metrics) (Metrics, error) {
	memorial.mutter.RLock() // <---- MUTEX
	defer memorial.mutter.RUnlock()
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
func (memorial MemoryStorageStruct) PutAllMetrics(ctx context.Context, metras *[]Metrics) error {
	memorial.mutter.Lock()
	defer memorial.mutter.Unlock()

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
func (memorial MemoryStorageStruct) GetAllMetrics(ctx context.Context) (*[]Metrics, error) {

	memorial.mutter.RLock()
	defer memorial.mutter.RUnlock()

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
