package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	sq "github.com/Masterminds/squirrel"
)

type Store struct {
	db *sqlx.DB
	sq sq.StatementBuilderType
}

func NewStore(db *sqlx.DB) (*Store, error) {
	sq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &Store{db: db, sq: sq}, nil
}

var now = func() time.Time {
	return time.Now().UTC().Round(time.Millisecond)
}

func (s *Store) SaveDeployment(ctx context.Context, def domain.AppDefinition) error {
	id := uuid.NewString()
	appPayload, err := json.Marshal(def.App)
	if err != nil {
		return fmt.Errorf("failed to marshal app definition to json: %w", err)
	}

	query, args, err := s.sq.Insert("deployments").
		Columns("id", "app_id", "app", "tag", "sha", "user", "created_at").
		Values(id, def.AppID, string(appPayload), def.Tag, def.Sha, def.User, now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SaveDeployment query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec SaveDeployment: %w", err)
	}

	return nil
}

func (s *Store) GetDeploymentHistory(ctx context.Context, appID string) ([]domain.AppDefinition, error) {
	query, args, err := s.sq.Select("id", "app_id", "app", "tag", "sha", "user", "created_at").
		From("deployments").
		Where(sq.Eq{"app_id": appID}).
		OrderBy("created_at DESC").
		Limit(20).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SaveDeployment query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetDeploymentHistory: %w", err)
	}
	defer rows.Close()

	var defs []domain.AppDefinition
	for rows.Next() {
		var def domain.AppDefinition
		var appPayload string
		if err := rows.Scan(&def.ID, &def.AppID, &appPayload, &def.Tag, &def.Sha, &def.User, &def.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan GetDeploymentHistory row: %w", err)
		}

		var a tqsdk.Space
		if err := json.Unmarshal([]byte(appPayload), &appPayload); err != nil {
			return nil, fmt.Errorf("failed to decode app payload in GetDeploymentHistory: %w", err)
		}
		def.App = a
		defs = append(defs, def)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occured in iterating GetDeploymentHistory rows: %w", err)
	}

	return defs, nil
}
