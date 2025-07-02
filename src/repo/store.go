package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/xid"

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
	id := xid.New().String()
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
	def.ID = xid.New().String()
	def.CreatedAt = now()
	appPayload, err := json.Marshal(def.Space)
	if err != nil {
		return def, fmt.Errorf("failed to marshal app definition to json: %w", err)
	}

	query, args, err := s.sq.Insert("deployments").
		Columns("id", "fromDeploymentId", "repoId", "space", "sha", "branch", "commitMessage", "buildTag", "userDisplayName", "status", "createdAt").
		Values(def.ID, def.FromDeploymentID, def.RepoID, string(appPayload), def.Sha, def.Branch, def.CommitMessage, def.BuildTag, def.UserDisplayName, def.Status, def.CreatedAt).
		ToSql()
	if err != nil {
		return def, fmt.Errorf("failed to build SaveDeployment query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return def, fmt.Errorf("failed to exec SaveDeployment: %w", err)
	}

	return def, nil
}

func (s *Store) UpdateDeployment(ctx context.Context, deployment domain.AppDeployment) error {
	appPayload, err := json.Marshal(deployment.Space)
	if err != nil {
		return fmt.Errorf("failed to marshal app definition to json: %w", err)
	}
	deployment.UpdatedAt = now()
	query, args, err := s.sq.Update("deployments").
		Set("space", appPayload).
		Set("sha", deployment.Sha).
		Set("branch", deployment.Branch).
		Set("commitMessage", deployment.CommitMessage).
		Set("buildTag", deployment.BuildTag).
		Set("status", deployment.Status).
		Where(sq.Eq{"id": deployment.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build UpdateDeploymentStatus query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec UpdateDeploymentStatus: %w", err)
	}

	return nil
}

func (s *Store) GetDeployment(ctx context.Context, deploymentID string) (domain.AppDeployment, error) {
	query, args, err := s.sq.Select("id", "fromDeploymentId", "repoId", "space", "sha", "branch", "commitMessage", "buildTag", "userDisplayName", "status", "createdAt").
		From("deployments").
		Where(sq.Eq{"id": deploymentID}).
		ToSql()
	if err != nil {
		return domain.AppDeployment{}, fmt.Errorf("failed to build GetDeployment query: %w", err)
	}

	var dep domain.AppDeployment
	var spacePayload string
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&dep.ID, &dep.FromDeploymentID, &dep.RepoID, &spacePayload, &dep.Sha, &dep.Branch, &dep.CommitMessage, &dep.BuildTag, &dep.UserDisplayName, &dep.Status, &dep.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dep, domain.ErrDeploymentNotFound
		}
		return dep, fmt.Errorf("failed to scan GetDeployment: %w", err)
	}

	var space tqsdk.Space
	if err := json.Unmarshal([]byte(spacePayload), &space); err != nil {
		return dep, fmt.Errorf("failed to unmarshal space in GetDeployment: %w", err)
	}
	dep.Space = space

	return dep, nil
}

func (s *Store) GetDeployments(ctx context.Context, repoID string) ([]domain.AppDeployment, error) {
	query, args, err := s.sq.Select("id", "fromDeploymentId", "repoId", "space", "sha", "branch", "commitMessage", "buildTag", "userDisplayName", "status", "createdAt").
		From("deployments").
		Where(sq.Eq{"repoId": repoID}).
		OrderBy("id DESC").
		Limit(20).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetDeploymentHistory query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetDeploymentHistory: %w", err)
	}
	defer rows.Close()

	var deps []domain.AppDeployment
	for rows.Next() {
		var dep domain.AppDeployment
		var spacePayload string
		if err := rows.Scan(&dep.ID, &dep.FromDeploymentID, &dep.RepoID, &spacePayload, &dep.Sha, &dep.Branch, &dep.CommitMessage, &dep.BuildTag, &dep.UserDisplayName, &dep.Status, &dep.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan GetDeploymentHistory row: %w", err)
		}

		var space tqsdk.Space
		if err := json.Unmarshal([]byte(spacePayload), &space); err != nil {
			return nil, fmt.Errorf("failed to decode app payload in GetDeploymentHistory: %w", err)
		}
		dep.Space = space
		deps = append(deps, dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occured in iterating GetDeploymentHistory rows: %w", err)
	}

	return deps, nil
}

// LinkGithub is called either on "created" app installation event
// or on manual sync,
// it cleans the previous installed repos state and inserts everything in the repos slice
func (s *Store) LinkGithub(ctx context.Context, installationID int, userDisplayName string, repos []domain.InstalledRepository) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to start link github transaction: %w", err)
	}
	defer tx.Rollback()

	userID, err := s.getUserIDByDisplayName(ctx, userDisplayName, s.db)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	treenqInstallationID, err := s.linkGithub(ctx, tx, installationID, userID, userDisplayName, repos)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit LinkGithub: %w", err)
	}
	return treenqInstallationID, nil
}

func (s *Store) linkGithub(ctx context.Context, querier Querier, installationID int, userID string, senderLogin string, repos []domain.InstalledRepository) (string, error) {
	timestamp := now()

	// save a github app
	treenqInstallationID, err := s.saveInstallation(ctx, querier, userID, installationID, timestamp)
	if err != nil {
		return "", fmt.Errorf("failed to save installation: %w", err)
	}

	// clean previous state
	err = s.removeInstalledRepos(ctx, userID, installationID, querier)
	if err != nil {
		return "", fmt.Errorf("failed to remove installed repos: %w", err)
	}

	// Insert repositories
	err = s.insertInstalledRepos(ctx, repos, userID, installationID, timestamp, querier)
	if err != nil {
		return "", fmt.Errorf("failed to save installed repos: %w", err)
	}

	return treenqInstallationID, nil
}

// LinkAllGithubInstallations saves all repositories from multiple installations for a user transactionally
func (s *Store) LinkAllGithubInstallations(ctx context.Context, profile domain.UserInfo, installationsRepos map[int][]domain.GithubRepository) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for installationID, repos := range installationsRepos {
		_, err := s.linkGithub(ctx, tx, installationID, profile.ID, profile.DisplayName, repos)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getUserIDByDisplayName gets user ID by display name
func (s *Store) getUserIDByDisplayName(ctx context.Context, senderLogin string, q Querier) (string, error) {
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

func (s *Store) removeInstalledRepos(
	ctx context.Context,
	userID string,
	installationID int,
	q Querier,
) error {
	repoQuery := s.sq.Delete("installedRepos").Where(sq.Eq{
		"userId":         userID,
		"installationId": installationID,
	})

	sql, args, err := repoQuery.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build repositories query: %w", err)
	}

	if _, err := q.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to insert repositories: %w", err)
	}

	return nil
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
		Columns("id", "githubId", "fullName", "private", "branch", "installationId", "userId", "status", "createdAt")

	for _, repo := range repos {
		id := xid.New().String()
		repoQuery = repoQuery.Values(
			id,
			repo.ID,
			repo.FullName,
			repo.Private,
			"",
			installationID,
			userID,
			domain.StatusRepoActive,
			createdAt,
		)
	}

	repoQuery = repoQuery.Suffix("ON CONFLICT (githubId) DO NOTHING")

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

func (s *Store) GetInstallationID(ctx context.Context, userID, repoName string) (int, error) {
	q, args, err := s.sq.Select("installationId").
		From("installedRepos").
		Where(sq.Eq{"userId": userID, "fullName": repoName}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build instalaltions query: %w", err)
	}

	var installationGithubID int
	if err := s.db.QueryRowContext(ctx, q, args...).Scan(&installationGithubID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrInstallationNotFound
		}
		return 0, fmt.Errorf("failed to get installation ID: %w", err)
	}
	return installationGithubID, nil
}

func (s *Store) SaveInstallation(ctx context.Context, userID string, githubID int) (string, error) {
	timestamp := now()

	return s.saveInstallation(ctx, s.db, userID, githubID, timestamp)
}

func (s *Store) saveInstallation(ctx context.Context, q Querier, userID string, githubID int, timestamp time.Time) (string, error) {
	// First check if installation already exists
	query, args, err := s.sq.Select("id").
		From("installations").
		Where(sq.Eq{
			"githubId": githubID,
			"userId":   userID,
		}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build check installation query: %w", err)
	}

	var installationID string
	err = q.QueryRowContext(ctx, query, args...).Scan(&installationID)
	if err == nil {
		// Installation found, return existing ID
		return installationID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrInstallationNotFound
	}

	// No existing installation found, create new one
	installationID = xid.New().String()

	query, args, err = s.sq.Insert("installations").
		Columns("id", "userId", "githubId", "createdAt").
		Values(installationID, userID, githubID, timestamp).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build save installation query: %w", err)
	}

	if _, err := q.ExecContext(ctx, query, args...); err != nil {
		return "", fmt.Errorf("failed to save installation: %w", err)
	}

	return installationID, nil
}

func (s *Store) GetGithubRepos(ctx context.Context, userID string) ([]domain.GithubRepository, bool, error) {
	hasInstallation, err := s.userHasInstallation(ctx, userID)
	if err != nil {
		return nil, false, nil
	}
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "status", "branch").
		From("installedRepos").
		Where(sq.Eq{"userId": userID}).
		OrderBy("id ASC").
		ToSql()
	if err != nil {
		return nil, hasInstallation, fmt.Errorf("failed to build GetGithubRepos query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, hasInstallation, fmt.Errorf("failed to query GetGithubRepos: %w", err)
	}
	defer rows.Close()

	var repos []domain.GithubRepository
	for rows.Next() {
		var repo domain.GithubRepository
		if err := rows.Scan(&repo.TreenqID, &repo.ID, &repo.FullName, &repo.Private, &repo.Status, &repo.Branch); err != nil {
			return nil, hasInstallation, fmt.Errorf("failed to scan GetGithubRepos row: %w", err)
		}

		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, hasInstallation, fmt.Errorf("failed to iterate GetGithubRepos rows: %w", err)
	}

	return repos, hasInstallation, nil
}

func (s *Store) userHasInstallation(ctx context.Context, userID string) (bool, error) {
	query, args, err := s.sq.Select("count(*)").
		From("installations").
		Where(sq.Eq{"userId": userID}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build userHasInstallation query: %w", err)
	}

	var count int
	row := s.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to scan userHasInstallation: %w", err)
	}

	return count > 0, nil
}

func (s *Store) ConnectRepo(ctx context.Context, userID, repoID, branch string, space tqsdk.Space) (domain.GithubRepository, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.GithubRepository{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	repo, err := s.connectRepo(ctx, tx, userID, repoID, branch)
	if err != nil {
		return repo, err
	}

	err = s.saveSpace(ctx, tx, repoID, space)
	if err != nil {
		return repo, err
	}

	if err := tx.Commit(); err != nil {
		return repo, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return repo, nil
}

func (s *Store) connectRepo(ctx context.Context, q Querier, userID, repoID, branch string) (domain.GithubRepository, error) {
	query, args, err := s.sq.Update("installedRepos").
		Set("branch", branch).
		Where(sq.Eq{"id": repoID, "userId": userID}).
		Suffix("RETURNING id, githubId, fullName, private, branch, status").
		ToSql()
	if err != nil {
		return domain.GithubRepository{}, fmt.Errorf("failed to build ConnectRepoBranch query: %w", err)
	}

	row := q.QueryRowContext(ctx, query, args...)
	if row.Err() != nil {
		return domain.GithubRepository{}, fmt.Errorf("failed to execute ConnectRepoBranch: %w", row.Err())
	}
	var repo domain.GithubRepository
	if err := row.Scan(&repo.TreenqID, &repo.ID, &repo.FullName, &repo.Private, &repo.Branch, &repo.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo, domain.ErrRepoNotFound
		}
		return repo, fmt.Errorf("failed to scan repository: %w", err)
	}
	return repo, nil
}

func (s *Store) SaveSpace(ctx context.Context, repoID string, space tqsdk.Space) error {
	return s.saveSpace(ctx, s.db, repoID, space)
}

func (s *Store) saveSpace(ctx context.Context, q Querier, repoID string, space tqsdk.Space) error {
	spacePayload, err := json.Marshal(space)
	if err != nil {
		return fmt.Errorf("failed to marshal space to json: %w", err)
	}

	timestamp := now()
	query, args, err := s.sq.Insert("spaces").
		Columns("repoId", "space", "createdAt", "updatedAt").
		Values(repoID, string(spacePayload), timestamp, timestamp).
		Suffix("ON CONFLICT (repoId) DO UPDATE SET space = EXCLUDED.space, updatedAt = EXCLUDED.updatedAt").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build saveSpace query: %w", err)
	}

	if _, err := q.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec saveSpace: %w", err)
	}

	return nil
}

func (s *Store) GetSpace(ctx context.Context, repoID string) (tqsdk.Space, error) {
	query, args, err := s.sq.Select("space").
		From("spaces").
		Where(sq.Eq{"repoId": repoID}).
		ToSql()
	if err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to build GetSpace query: %w", err)
	}

	var spacePayload string
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&spacePayload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tqsdk.Space{}, domain.ErrNoSpaceFound
		}
		return tqsdk.Space{}, fmt.Errorf("failed to scan GetSpace: %w", err)
	}

	var space tqsdk.Space
	if err := json.Unmarshal([]byte(spacePayload), &space); err != nil {
		return tqsdk.Space{}, domain.ErrTqIsNotValidJson
	}

	return space, nil
}

func (s *Store) RepoIsConnected(ctx context.Context, repoID string) (bool, error) {
	query, args, err := s.sq.Select("branch").
		From("installedRepos").
		Where(sq.Eq{"id": repoID}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build RepoIsConnected query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var branch string
	if err := row.Scan(&branch); err != nil {
		return false, fmt.Errorf("failed to scan RepoIsConnected value: %w", err)
	}

	return branch != "", nil
}

func (s *Store) GetRepoByGithub(ctx context.Context, githubRepoID int) (domain.GithubRepository, error) {
	var repo domain.GithubRepository
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "branch", "installationId", "status").
		From("installedRepos").
		Where(sq.Eq{"githubId": githubRepoID}).
		ToSql()
	if err != nil {
		return repo, fmt.Errorf("failed to build GetRepoByGithub query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&repo.TreenqID, &repo.ID, &repo.FullName,
		&repo.Private, &repo.Branch, &repo.InstallationID, &repo.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.GithubRepository{}, domain.ErrRepoNotFound
		}
		return domain.GithubRepository{}, fmt.Errorf("failed to scan GetRepoByGithub value: %w", err)
	}

	return repo, nil
}

func (s *Store) GetRepoByID(ctx context.Context, userID string, repoID string) (domain.GithubRepository, error) {
	var repo domain.GithubRepository
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "branch", "installationId", "status").
		From("installedRepos").
		Where(sq.Eq{"id": repoID, "userId": userID}).
		ToSql()
	if err != nil {
		return repo, fmt.Errorf("failed to build GetRepoByID query: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&repo.TreenqID, &repo.ID, &repo.FullName,
		&repo.Private, &repo.Branch, &repo.InstallationID, &repo.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo, domain.ErrRepoNotFound
		}

		return domain.GithubRepository{}, fmt.Errorf("failed to scan GetRepoByID value: %w", err)
	}

	return repo, nil
}

func (s *Store) SaveSecret(ctx context.Context, repoID, key, userDisplayName string) error {
	createdAt := now()
	query, args, err := s.sq.Insert("secrets").
		Columns("repoId", "key", "userDisplayName", "createdAt").
		Values(repoID, key, userDisplayName, createdAt).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SaveSecret query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec SaveSecret: %w", err)
	}

	return nil
}

func (s *Store) GetRepositorySecretKeys(ctx context.Context, repoID, userDisplayName string) ([]string, error) {
	query, args, err := s.sq.Select("key").
		From("secrets").
		Where(sq.Eq{"repoId": repoID, "userDisplayName": userDisplayName}).
		OrderBy("createdAt ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetRepositorySecretKeys query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetRepositorySecretKeys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan GetRepositorySecretKeys row: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred while iterating GetRepositorySecretKeys rows: %w", err)
	}

	return keys, nil
}

func (s *Store) RepositorySecretKeyExists(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
	query, args, err := s.sq.Select("1").
		From("secrets").
		Where(sq.Eq{"repoId": repoID, "userDisplayName": userDisplayName, "key": key}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build RepositorySecretKeyExists query: %w", err)
	}

	var exists int
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to query RepositorySecretKeyExists: %w", err)
	}

	return true, nil
}

func (s *Store) RemoveSecret(ctx context.Context, repoID, key, userDisplayName string) error {
	query, args, err := s.sq.Delete("secrets").
		Where(sq.Eq{"repoId": repoID, "key": key, "userDisplayName": userDisplayName}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build RemoveSecret query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec RemoveSecret: %w", err)
	}

	return nil
}
