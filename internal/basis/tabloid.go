package basis

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"gorono/internal/models"
)

type Metrics = models.Metrics

type DBstruct struct {
}

func TableCreation(ctx context.Context, db *pgx.Conn) error {
	crea := "CREATE TABLE IF NOT EXISTS Gauge(metricname VARCHAR(50) PRIMARY KEY, value FLOAT8);"
	tag, err := db.Exec(ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Gauge table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	crea = "CREATE TABLE IF NOT EXISTS Counter(metricname VARCHAR(50) PRIMARY KEY, value BIGINT);"
	tag, err = db.Exec(ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Counter table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	return nil
}

// -------------- put ONE metric to the table
func (gag DBstruct) PutMetric(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct, metr *Metrics) error {
	//func (dataBase models.DBstruct) TableUpSert(ctx context.Context, db *pgx.Conn, metr *Metrics) error {
	if (metr.MType == "gauge" && metr.Value == nil) ||
		(metr.MType == "counter" && metr.Delta == nil) ||
		(metr.Value != nil && metr.Delta != nil) {
		return fmt.Errorf("wrong metric %+v", metr)
	}
	var order string
	switch metr.MType {
	case "gauge":
		order = fmt.Sprintf("INSERT INTO Gauge AS args(metricname, value) VALUES ('%[1]s',%[2]g) ", metr.ID, *metr.Value)
		order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=EXCLUDED.value;"
	case "counter":
		order = fmt.Sprintf("INSERT INTO Counter AS args(metricname, value) VALUES ('%[1]s',%[2]d) ", metr.ID, *metr.Delta)
		order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=args.value+EXCLUDED.value;"
		// args.value - старое значение. EXCLUDED.value - новое, переданное для вставки или обновления
	default:
		return fmt.Errorf("wrong metric type \"%s\"", metr.MType)
	}
	_, err := db.Exec(ctx, order)
	if err != nil {
		return fmt.Errorf("error insert/update %+v error is %w", metr, err)
	}
	return nil
}

// ------ get ONE metric from the table
func (gag DBstruct) GetMetric(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct, metr *Metrics) (Metrics, error) {
	//func TableGetMetric(ctx context.Context, db *pgx.Conn, metr *Metrics) error {
	switch metr.MType {
	case "gauge":
		var flo float64 // here we scan Value
		order := "SELECT value FROM gauge WHERE metricname = $1;"
		row := db.QueryRow(ctx, order, metr.ID)
		err := row.Scan(&flo)
		if err != nil {
			return *metr, fmt.Errorf("error get %s gauge metric.  %w", metr.ID, err)
		}
		metr.Value = &flo
	case "counter":
		var inta int64 // here we scan Delta
		order := "SELECT value FROM counter WHERE metricname = $1;"
		row := db.QueryRow(ctx, order, metr.ID)
		err := row.Scan(&inta)
		if err != nil {
			return *metr, fmt.Errorf("error get %s counter metric.  %w", metr.ID, err)
		}
		metr.Delta = &inta
	default:
		return *metr, fmt.Errorf("wrong metric type \"%s\"", metr.MType)
	}
	return *metr, nil
}

// ----------- transaction. PUT ALL metrics to the tables ----------------------
func (gag DBstruct) PutAllMetrics(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct, metras *[]Metrics) error {
	//func TableBuncher(ctx context.Context, db *pgx.Conn, metras *[]Metrics) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error db.Begin  %[1]w", err)
	}
	var order string
	for _, metr := range *metras {
		if (metr.MType == "gauge" && metr.Value == nil) || (metr.MType == "counter" && metr.Delta == nil) {
			log.Printf("wrong metric %+v", metr)
			continue
		}
		switch metr.MType {
		case "gauge":
			order = fmt.Sprintf("INSERT INTO Gauge AS args(metricname, value) VALUES ('%[1]s',%[2]g) ", metr.ID, *metr.Value)
			order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=EXCLUDED.value;"
		case "counter":
			order = fmt.Sprintf("INSERT INTO Counter AS args(metricname, value) VALUES ('%[1]s',%[2]d) ", metr.ID, *metr.Delta)
			order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=args.value+EXCLUDED.value;"
			// args.value - старое значение. EXCLUDED.value - новое, переданное для вставки или обновления
		default:
			log.Printf("wrong metric type \"%s\"\n", metr.MType)
			continue
		}
		_, err := tx.Exec(ctx, order)
		if err != nil {
			log.Printf("error put %+v. error is %v", metr, err)
		}
	}
	return tx.Commit(ctx)
}

// ------- get ALL metrics from the tables
func (gag DBstruct) GetAllMetrics(ctx context.Context, db *pgx.Conn, memorial *models.MemoryStorageStruct) (*[]Metrics, error) {
	//func TableGetAllTables(ctx context.Context, db *pgx.Conn, metras *[]Metrics) error {
	zapros := `select 'counter' AS metrictype, metricname AS name, null AS value, value AS delta from counter
		UNION
	select 'gauge' AS metrictype, metricname as name, value as value, null as delta from gauge`

	var inta int64
	var flo float64
	metr := Metrics{ID: "", MType: "", Value: &flo, Delta: &inta}

	rows, err := db.Query(ctx, zapros)
	if err != nil {
		return nil, fmt.Errorf("error Query %[2]s:%[3]d  %[1]w", err, db.Config().Host, db.Config().Port)
	}
	metras := []Metrics{}
	for rows.Next() {
		err = rows.Scan(&metr.MType, &metr.ID, &metr.Value, &metr.Delta)
		if err != nil {
			return nil, fmt.Errorf("error table Scan %[2]s:%[3]d  %[1]w", err, db.Config().Host, db.Config().Port)
		}
		metras = append(metras, metr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("err := rows.Err()  %w", err)
	}
	return &metras, nil
}
