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

type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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

func (s *Store) SaveDeployment(ctx context.Context, def domain.AppDeployment) (domain.AppDeployment, error) {
	def.ID = uuid.NewString()
	def.CreatedAt = now()
	appPayload, err := json.Marshal(def.Space)
	if err != nil {
		return def, fmt.Errorf("failed to marshal app definition to json: %w", err)
	}

	query, args, err := s.sq.Insert("deployments").
		Columns("id", "repoId", "space", "sha", "buildTag", "userDisplayName", "createdAt").
		Values(def.ID, def.RepoID, string(appPayload), def.Sha, def.UserDisplayName, def.CreatedAt).
		ToSql()
	if err != nil {
		return def, fmt.Errorf("failed to build SaveDeployment query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return def, fmt.Errorf("failed to exec SaveDeployment: %w", err)
	}

	return def, nil
}

func (s *Store) GetDeploymentHistory(ctx context.Context, repoID string) ([]domain.AppDeployment, error) {
	query, args, err := s.sq.Select("id", "repoId", "space", "sha", "buildTag", "userDisplayName", "createdAt").
		From("deployments").
		Where(sq.Eq{"id": repoID}).
		OrderBy("createdAt DESC").
		Limit(20).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetDeploymentHistory query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetDeploymentHistory: %w", err)
	}
	defer rows.Close()

	var deps []domain.AppDeployment
	for rows.Next() {
		var dep domain.AppDeployment
		var spacePayload string
		if err := rows.Scan(&dep.ID, &dep.RepoID, &spacePayload, &dep.Sha, &dep.BuildTag, &dep.UserDisplayName, &dep.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan GetDeploymentHistory row: %w", err)
		}

		var a tqsdk.Space
		if err := json.Unmarshal([]byte(spacePayload), &spacePayload); err != nil {
			return nil, fmt.Errorf("failed to decode app payload in GetDeploymentHistory: %w", err)
		}
		dep.Space = a
		deps = append(deps, dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occured in iterating GetDeploymentHistory rows: %w", err)
	}

	return deps, nil
}

func (s *Store) LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []domain.InstalledRepository) error {
	userID, err := s.getUserIDByDisplayName(ctx, senderLogin, s.db)
	if err != nil {
		return err
	}

	timestamp := now()

	// Insert repositories
	err = s.insertInstalledRepos(ctx, repos, userID, installationID, timestamp, s.db)
	if err != nil {
		return err
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
	userID string,
	installationID int,
	createdAt time.Time,
	q Querier,
) error {
	if len(repos) == 0 {
		return nil
	}

	repoQuery := s.sq.Insert("installedRepos").
		Columns("id", "githubId", "fullName", "private", "installationId", "userId", "status", "connected", "createdAt")

	for _, repo := range repos {
		id := uuid.NewString()
		repoQuery = repoQuery.Values(
			id,
			repo.ID,
			repo.FullName,
			repo.Private,
			installationID,
			userID,
			domain.StatusRepoActive,
			false,
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

	timestamp := now()
	err = s.insertInstalledRepos(ctx, repos, userID, installationID, timestamp, s.db)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) RemoveGithubRepos(ctx context.Context, installationID int, repos []domain.InstalledRepository) error {
	if len(repos) == 0 {
		return nil
	}
	repoIDs := make([]int, len(repos))
	for i, repo := range repos {
		repoIDs[i] = repo.ID
	}

	deleteQuery, args, err := s.sq.Delete("installedRepos").
		Where(sq.And{
			sq.Eq{"installationId": installationID},
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

func (s *Store) GetGithubRepos(ctx context.Context, userID string) ([]domain.Repository, error) {
	query, args, err := s.sq.Select("r.id", "r.githubId", "r.fullName", "r.private", "r.status", "r.connected").
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

	var repos []domain.Repository
	for rows.Next() {
		var repo domain.Repository
		if err := rows.Scan(&repo.TreenqID, &repo.ID, &repo.FullName, &repo.Private, &repo.Status, &repo.Connected); err != nil {
			return nil, fmt.Errorf("failed to scan GetGithubRepos row: %w", err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate GetGithubRepos rows: %w", err)
	}

	return repos, nil
}

func (s *Store) ConnectRepo(ctx context.Context, userID, repoID string) (domain.Repository, error) {
	query, args, err := s.sq.Update("installedRepos").
		Set("connected", true).
		Where(sq.Eq{"id": repoID, "userId": userID}).
		Suffix("RETURNING id, githubId, fullName, private,  status, connected").
		ToSql()
	if err != nil {
		return domain.Repository{}, fmt.Errorf("failed to build ConnectRepoBranch query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	if err != nil {
		return domain.Repository{}, fmt.Errorf("failed to execute ConnectRepoBranch: %w", err)
	}
	var repo domain.Repository
	if err := row.Scan(&repo.TreenqID, &repo.ID, &repo.FullName, &repo.Private, &repo.Status, &repo.Connected); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo, domain.ErrRepoNotFound
		}
	}
	return repo, nil
}

func (s *Store) RepoIsConnected(ctx context.Context, repoID string) (bool, error) {
	query, args, err := s.sq.Select("connected").
		From("installedRepos").
		Where(sq.Eq{"id": repoID}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build RepoIsConnected query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var connected bool
	if err := row.Scan(&connected); err != nil {
		return connected, fmt.Errorf("failed to scan RepoIsConnected value: %w", err)
	}

	return connected, nil
}

func (s *Store) GetRepoByGithub(ctx context.Context, githubRepoID int) (domain.Repository, error) {
	var repo domain.Repository
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "installationId", "status", "connected").
		From("installedRepos").
		Where(sq.Eq{"githubId": githubRepoID}).
		ToSql()
	if err != nil {
		return repo, fmt.Errorf("failed to build GetRepoByGithub query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&repo.TreenqID, &repo.ID, &repo.FullName,
		&repo.Private, &repo.InstallationID, &repo.Status, &repo.Connected); err != nil {
		return domain.Repository{}, fmt.Errorf("failed to scan GetRepoByGithub value: %w", err)
	}

	return repo, nil
}

func (s *Store) GetRepoByID(ctx context.Context, userID string, repoID string) (domain.Repository, error) {
	var repo domain.Repository
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "installationId", "status", "connected").
		From("installedRepos").
		Where(sq.Eq{"id": repoID, "userId": userID}).
		ToSql()
	if err != nil {
		return repo, fmt.Errorf("failed to build GetRepoByID query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&repo.TreenqID, &repo.ID, &repo.FullName,
		&repo.Private, &repo.InstallationID, &repo.Status, &repo.Connected); err != nil {
		return domain.Repository{}, fmt.Errorf("failed to scan GetRepoByID value: %w", err)
	}

	return repo, nil
}
