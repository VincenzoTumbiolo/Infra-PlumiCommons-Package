package s3migrator

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type MigrateAction string

const (
	UploadAction MigrateAction = "UPLOAD"
	DeleteAction MigrateAction = "DELETE"
)

type MigrateStmt interface {
	Action() MigrateAction
}

func UnmarshalMigrationJSON(data []byte) ([]MigrateStmt, error) {
	var rawStmts []map[string]string
	if err := json.Unmarshal(data, &rawStmts); err != nil {
		return nil, err
	}

	stmts := make([]MigrateStmt, 0, len(rawStmts))
	for _, rawStmt := range rawStmts {
		action, ok := rawStmt["action"]
		if !ok {
			return nil, fmt.Errorf("missing action field in migration statement")
		}

		switch MigrateAction(action) {
		case UploadAction:
			if len(rawStmt) != 3 {
				slog.Warn("Raw UPLOAD statement has an unexpected schema", "rawStmt", rawStmt)
			}

			path, ok := rawStmt["path"]
			if !ok {
				return nil, fmt.Errorf("missing path field in upload migration statement")
			}

			filename, ok := rawStmt["filename"]
			if !ok {
				return nil, fmt.Errorf("missing filename field in upload migration statement")
			}

			stmts = append(stmts, UploadStmt{
				Path:     path,
				Filename: filename,
			})

		case DeleteAction:
			if len(rawStmt) != 2 {
				slog.Warn("Raw DELETE statement has an unexpected schema", "rawStmt", rawStmt)
			}

			filename, ok := rawStmt["filename"]
			if !ok {
				return nil, fmt.Errorf("missing filename field in upload migration statement")
			}

			stmts = append(stmts, DeleteStmt{
				Filename: filename,
			})

		default:
			return nil, fmt.Errorf("unknown action %s in migration statement", action)
		}
	}

	return stmts, nil
}

type UploadStmt struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
}

func (UploadStmt) Action() MigrateAction {
	return UploadAction
}

type DeleteStmt struct {
	Filename string `json:"filename"`
}

func (DeleteStmt) Action() MigrateAction {
	return DeleteAction
}

type MigrationState struct {
	Version int  `json:"version"`
	Dirty   bool `json:"dirty"`
}
