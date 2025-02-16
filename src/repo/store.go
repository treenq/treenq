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

type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (s *Store) LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []domain.InstalledRepository) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for LinkGithub: %w", err)
	}
	defer tx.Rollback()

	userID, err := s.getUserIDByDisplayName(ctx, senderLogin, tx)
	if err != nil {
		return err
	}

	installationInternalID := uuid.NewString()
	timestamp := now()

	// Insert installation record
	installQuery, args, err := s.sq.Insert("installations").
		Columns("id", "githubId", "userId", "status", "createdAt", "updatedAt").
		Values(installationInternalID, installationID, userID, "active", timestamp, timestamp).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build installation query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, installQuery, args...); err != nil {
		return fmt.Errorf("failed to insert installation: %w", err)
	}

	// Insert repositories
	err = s.insertInstalledRepos(ctx, repos, userID, installationInternalID, timestamp, tx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Store) getUserIDByDisplayName(ctx context.Context, senderLogin string, q Querier) (string, error) {
	// Get user ID by display name
	userQuery, userArgs, err := s.sq.Select("id").
		From("users").
		Where(sq.Eq{"displayName": senderLogin}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build user query: %w", err)
	}

	var userID string
	if err := q.QueryRowContext(ctx, userQuery, userArgs...).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", fmt.Errorf("failed to get user ID: %w", err)
	}
	return userID, nil
}

func (s *Store) insertInstalledRepos(
	ctx context.Context,
	repos []domain.InstalledRepository,
	userID, installationInternalID string,
	createdAt time.Time,
	q Querier,
) error {
	if len(repos) == 0 {
		return nil
	}

	repoQuery := s.sq.Insert("installedRepos").
		Columns("id", "githubId", "fullName", "private", "installationId", "userId", "branch", "createdAt")

	for _, repo := range repos {
		id := uuid.NewString()
		repoQuery = repoQuery.Values(
			id,
			repo.ID,
			repo.FullName,
			repo.Private,
			installationInternalID,
			userID,
			"",
			createdAt,
		)
	}

	sql, args, err := repoQuery.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build repositories query: %w", err)
	}

	if _, err := q.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to insert repositories: %w", err)
	}

	return nil
}

func (s *Store) SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []domain.InstalledRepository) error {
	userID, err := s.getUserIDByDisplayName(ctx, senderLogin, s.db)
	if err != nil {
		return err
	}

	installationInternalID, err := s.getInstallationByGithub(ctx, installationID)
	if err != nil {
		return err
	}
	timestamp := now()

	err = s.insertInstalledRepos(ctx, repos, userID, installationInternalID, timestamp, s.db)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) getInstallationByGithub(ctx context.Context, githubID int) (string, error) {
	userQuery, userArgs, err := s.sq.Select("id").
		From("installations").
		Where(sq.Eq{"githubId": githubID}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build user query: %w", err)
	}

	var installationID string
	if err := s.db.QueryRowContext(ctx, userQuery, userArgs...).Scan(&installationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", fmt.Errorf("failed to get user ID: %w", err)
	}
	return installationID, nil
}

func (s *Store) RemoveGithubRepos(ctx context.Context, installationID int, repos []domain.InstalledRepository) error {
	if len(repos) == 0 {
		return nil
	}
	installationInternalID, err := s.getInstallationByGithub(ctx, installationID)
	if err != nil {
		return err
	}

	repoIDs := make([]int, len(repos))
	for i, repo := range repos {
		repoIDs[i] = repo.ID
	}

	deleteQuery, args, err := s.sq.Delete("installedRepos").
		Where(sq.And{
			sq.Eq{"installationId": installationInternalID},
			sq.Eq{"githubId": repoIDs},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, deleteQuery, args...); err != nil {
		return fmt.Errorf("failed to delete repositories: %w", err)
	}

	return nil
}

func (s *Store) GetGithubRepos(ctx context.Context, userID string) ([]domain.InstalledRepository, error) {
	query, args, err := s.sq.Select("r.id", "r.githubId", "r.fullName", "r.private", "r.branch").
		From("installedRepos r").
		Where(sq.Eq{"r.userId": userID}).
		OrderBy("r.createdAt ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetGithubRepos query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetGithubRepos: %w", err)
	}
	defer rows.Close()

	var repos []domain.InstalledRepository
	for rows.Next() {
		var repo domain.InstalledRepository
		if err := rows.Scan(&repo.TreenqID, &repo.ID, &repo.FullName, &repo.Private, &repo.Branch); err != nil {
			return nil, fmt.Errorf("failed to scan GetGithubRepos row: %w", err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate GetGithubRepos rows: %w", err)
	}

	return repos, nil
}

func (s *Store) ConnectRepoBranch(ctx context.Context, repoID int, branch string) error {
	query, args, err := s.sq.Update("installedRepos").
		Set("branch", branch).
		Where(sq.Eq{"githubId": repoID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build ConnectRepoBranch query: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute ConnectRepoBranch: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no repository found with ID %d", repoID)
	}

	return nil
}
