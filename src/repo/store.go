package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

var ErrNoAuthState = fmt.Errorf("no auth state found")

func (s *Store) GetOrCreateUser(ctx context.Context, user domain.UserInfo) (domain.UserInfo, error) {
	query, args, err := s.sq.Select("id").From("users").Where(sq.Eq{"email": user.Email}).ToSql()
	if err != nil {
		return user, fmt.Errorf("failed to build select query GetOrCreateUser: %w", err)
	}
	row := s.db.QueryRow(query, args...)
	if err := row.Scan(&user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.createUser(ctx, user)
		}

		return user, fmt.Errorf("failed to scan select GetOrCreateUser: %w", err)
	}

	return user, nil
}

func (s *Store) createUser(ctx context.Context, user domain.UserInfo) (domain.UserInfo, error) {
	id := uuid.NewString()
	query, args, err := s.sq.Insert("users").
		Columns("id", "email", "displayName").
		Values(id, user.Email, user.DisplayName).
		ToSql()
	if err != nil {
		return user, fmt.Errorf("failed to build query createUser: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return user, fmt.Errorf("failed to exec createUser: %w", err)
	}

	user.ID = id
	return user, nil
}

func (s *Store) SaveDeployment(ctx context.Context, def domain.AppDefinition) (domain.AppDefinition, error) {
	id := uuid.NewString()
	def.ID = id
	appPayload, err := json.Marshal(def.App)
	if err != nil {
		return def, fmt.Errorf("failed to marshal app definition to json: %w", err)
	}

	query, args, err := s.sq.Insert("deployments").
		Columns("id", "appId", "app", "tag", "sha", "user", "createdAt").
		Values(id, def.AppID, string(appPayload), def.Tag, def.Sha, def.User, now()).
		ToSql()
	if err != nil {
		return def, fmt.Errorf("failed to build SaveDeployment query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return def, fmt.Errorf("failed to exec SaveDeployment: %w", err)
	}

	return def, nil
}

func (s *Store) GetDeploymentHistory(ctx context.Context, appID string) ([]domain.AppDefinition, error) {
	query, args, err := s.sq.Select("id", "appId", "app", "tag", "sha", "user", "createdAt").
		From("deployments").
		Where(sq.Eq{"appId": appID}).
		OrderBy("createdAt DESC").
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

func (s *Store) LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []domain.InstalledRepository) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for LinkGithub: %w", err)
	}
	defer tx.Rollback()

	// Insert installation record
	installQuery, args, err := s.sq.Insert("installations").
		Columns("githubId", "orgName", "status", "createdAt").
		Values(installationID, senderLogin, "active", now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build installation query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, installQuery, args...); err != nil {
		return fmt.Errorf("failed to insert installation: %w", err)
	}

	// Insert repositories
	if len(repos) > 0 {
		repoQuery := s.sq.Insert("installedRepos").
			Columns("githubId", "fullName", "private", "installationId", "createdAt")

		for _, repo := range repos {
			repoQuery = repoQuery.Values(
				repo.ID,
				repo.FullName,
				repo.Private,
				installationID,
				now(),
			)
		}

		sql, args, err := repoQuery.ToSql()
		if err != nil {
			return fmt.Errorf("failed to build repositories query: %w", err)
		}

		if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
			return fmt.Errorf("failed to insert repositories: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Store) GetConnectedRepositories(ctx context.Context, email string) ([]domain.GithubRepository, error) {
	query, args, err := s.sq.Select("id", "fullName").
		From("repos").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetConnectedRepositories query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetConnectedRepositories: %w", err)
	}
	defer rows.Close()

	var repos []domain.GithubRepository
	for rows.Next() {
		var repo domain.GithubRepository
		if err := rows.Scan(&repo.ID, &repo.FullName); err != nil {
			return nil, fmt.Errorf("failed to scan GetConnectedRepositories row: %w", err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occured in iterating GetConnectedRepositories rows: %w", err)
	}

	return repos, nil
}

func (s *Store) SaveConnectedRepository(ctx context.Context, email string, repo domain.GithubRepository) error {
	query, args, err := s.sq.Insert("repos").
		Columns("id", "fullName", "email", "createdAt").
		Values(repo.ID, repo.FullName, email, now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SaveConnectedRepository query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec SaveConnectedRepository: %w", err)
	}

	return nil
}

func (s *Store) RemoveConnectedRepository(ctx context.Context, email string, repoID int) error {
	query, args, err := s.sq.Delete("repos").
		Where(sq.Eq{"id": repoID, "email": email}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build DisconnectRepository query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec DisconnectRepository: %w", err)
	}

	return nil
}

func (s *Store) SaveGithubRepos(ctx context.Context, email string, repos []domain.GithubRepository) error {
	if len(repos) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for SaveGithubRepos: %w", err)
	}
	defer tx.Rollback()

	deleteQuery, args, err := s.sq.Delete("githubRepos").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build deleting SaveGithubRepos query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, deleteQuery, args...); err != nil {
		return fmt.Errorf("failed to exec deleting SaveGithubRepos: %w", err)
	}

	query := s.sq.Insert("githubRepos").
		Columns("email", "repoId", "fullName", "defaultBranch", "createdAt")

	for _, repo := range repos {
		query = query.Values(email, repo.ID, repo.FullName, repo.DefaultBranch, now())
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SaveGithubRepos query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to exec SaveGithubRepos: %w", err)
	}

	return nil
}

func (s *Store) GetGithubRepos(ctx context.Context, email string) ([]domain.GithubRepository, error) {
	query, args, err := s.sq.Select("repoId", "fullName", "defaultBranch").
		From("githubRepos").
		Where(sq.Eq{"email": email}).
		OrderBy("createdAt DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetGithubRepos query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetGithubRepos: %w", err)
	}
	defer rows.Close()

	var repos []domain.GithubRepository
	for rows.Next() {
		var repo domain.GithubRepository
		if err := rows.Scan(&repo.ID, &repo.FullName, &repo.DefaultBranch); err != nil {
			return nil, fmt.Errorf("failed to scan GetGithubRepos row: %w", err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred in iterating GetGithubRepos rows: %w", err)
	}

	return repos, nil
}
