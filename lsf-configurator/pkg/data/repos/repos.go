package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"lsf-configurator/pkg/core"
	"sync"
)

var dbWriteMutex sync.Mutex

type functionAppRepo struct {
	db *sql.DB
}

func NewFunctionAppRepository(db *sql.DB) core.FunctionAppRepository {
	return &functionAppRepo{db: db}
}

func (r *functionAppRepo) Save(app *core.FunctionApp) error {
	dbWriteMutex.Lock()
	defer dbWriteMutex.Unlock()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT OR REPLACE INTO function_apps (id, name) VALUES (?, ?)`, app.Id, app.Name)
	if err != nil {
		return err
	}

	compRepo := NewFunctionCompositionRepository(r.db).(*functionCompositionRepo)

	// Save each composition
	for _, comp := range app.Compositions {
		err := compRepo.saveWithTx(tx, comp)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *functionAppRepo) GetByID(id string) (*core.FunctionApp, error) {
	row := r.db.QueryRow(`SELECT id, name FROM function_apps WHERE id = ?`, id)

	var app core.FunctionApp
	if err := row.Scan(&app.Id, &app.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	app.Compositions = make([]*core.FunctionComposition, 0)

	rows, err := r.db.Query(`SELECT id FROM function_compositions WHERE function_app_id = ?`, app.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	compRepo := NewFunctionCompositionRepository(r.db)

	for rows.Next() {
		var compID string
		if err := rows.Scan(&compID); err != nil {
			return nil, err
		}
		comp, err := compRepo.GetByID(compID)
		if err != nil {
			return nil, err
		}
		app.Compositions = append(app.Compositions, comp)
	}

	return &app, nil
}

func (r *functionAppRepo) Delete(id string) error {
	dbWriteMutex.Lock()
	defer dbWriteMutex.Unlock()
	_, err := r.db.Exec(`DELETE FROM function_apps WHERE id = ?`, id)
	return err
}

type functionCompositionRepo struct {
	db *sql.DB
}

func NewFunctionCompositionRepository(db *sql.DB) core.FunctionCompositionRepository {
	return &functionCompositionRepo{db: db}
}

func (r *functionCompositionRepo) Save(comp *core.FunctionComposition) error {
	dbWriteMutex.Lock()
	defer dbWriteMutex.Unlock()
	filesJSON, err := json.Marshal(comp.Files)
	if err != nil {
		return fmt.Errorf("failed to marshal files: %w", err)
	}

	componentsJSON, err := json.Marshal(comp.Components)
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT OR REPLACE INTO function_compositions (
			id, function_app_id, node, namespace, source_path, runtime,
			image, timestamp, files, components
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		comp.Id, comp.FunctionAppId, comp.Node, comp.NameSpace, comp.SourcePath,
		comp.Runtime, comp.Image, comp.Timestamp, string(filesJSON), string(componentsJSON),
	)

	return err
}

func (r *functionCompositionRepo) GetByID(id string) (*core.FunctionComposition, error) {
	row := r.db.QueryRow(`
		SELECT id, function_app_id, node, namespace, source_path, runtime,
		       image, timestamp, files, components
		FROM function_compositions
		WHERE id = ?`, id)

	var comp core.FunctionComposition
	var filesJSON, componentsJSON string

	err := row.Scan(
		&comp.Id, &comp.FunctionAppId, &comp.Node, &comp.NameSpace, &comp.SourcePath,
		&comp.Runtime, &comp.Image, &comp.Timestamp,
		&filesJSON, &componentsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(filesJSON), &comp.Files); err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}
	if err := json.Unmarshal([]byte(componentsJSON), &comp.Components); err != nil {
		return nil, fmt.Errorf("failed to parse components: %w", err)
	}

	return &comp, nil
}

func (r *functionCompositionRepo) Delete(id string) error {
	dbWriteMutex.Lock()
	defer dbWriteMutex.Unlock()
	_, err := r.db.Exec(`DELETE FROM function_compositions WHERE id = ?`, id)
	return err
}

func (r *functionCompositionRepo) saveWithTx(tx *sql.Tx, comp *core.FunctionComposition) error {
	filesJSON, err := json.Marshal(comp.Files)
	if err != nil {
		return fmt.Errorf("failed to marshal files: %w", err)
	}
	componentsJSON, err := json.Marshal(comp.Components)
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO function_compositions (
			id, function_app_id, node, namespace, source_path, runtime,
			image, timestamp, files, components
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		comp.Id, comp.FunctionAppId, comp.Node, comp.NameSpace,
		comp.SourcePath, comp.Runtime, comp.Image, comp.Timestamp,
		string(filesJSON), string(componentsJSON),
	)
	return err
}
