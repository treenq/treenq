package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/rs/xid"

	"github.com/jmoiron/sqlx"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	sq "github.com/Masterminds/squirrel"
)

const personalWorkspaceName = "default"

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

	// Get user's workspaces
	workspaces, err := s.GetUserWorkspaces(ctx, user.ID)
	if err != nil {
		return user, fmt.Errorf("failed to get user workspaces: %w", err)
	}

	// Extract workspace IDs
	workspaceIDs := make([]string, len(workspaces))
	for i, workspace := range workspaces {
		workspaceIDs[i] = workspace.ID
	}
	user.Workspaces = workspaceIDs

	return user, nil
}

func (s *Store) createUser(ctx context.Context, user domain.UserInfo) (domain.UserInfo, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return user, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	id := xid.New().String()
	query, args, err := s.sq.Insert("users").
		Columns("id", "email", "displayName").
		Values(id, user.Email, user.DisplayName).
		ToSql()
	if err != nil {
		return user, fmt.Errorf("failed to build query createUser: %w", err)
	}
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return user, fmt.Errorf("failed to exec createUser: %w", err)
	}

	workspace, err := s.createDefaultWorkspaceForUser(ctx, tx, id)
	if err != nil {
		return user, err
	}

	if err := tx.Commit(); err != nil {
		return user, fmt.Errorf("failed to commit transaction: %w", err)
	}

	user.ID = id
	user.Workspaces = []string{workspace.ID}
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

func (s *Store) GetDeployment(ctx context.Context, workspaceID, deploymentID string) (domain.AppDeployment, error) {
	query, args, err := s.sq.Select("d.id", "d.fromDeploymentId", "d.repoId", "d.space", "d.sha", "d.branch", "d.commitMessage",
		"d.buildTag", "d.userDisplayName", "d.status", "d.createdAt", "d.updatedAt").
		From("deployments d").
		Join("installedRepos r ON d.repoId = r.id").
		Where(sq.And{
			sq.Eq{"d.id": deploymentID},
			sq.Eq{"r.workspaceId": workspaceID},
		}).
		ToSql()
	if err != nil {
		return domain.AppDeployment{}, fmt.Errorf("failed to build GetDeployment query: %w", err)
	}

	var dep domain.AppDeployment
	var spacePayload string
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&dep.ID, &dep.FromDeploymentID, &dep.RepoID, &spacePayload, &dep.Sha, &dep.Branch, &dep.CommitMessage, &dep.BuildTag, &dep.UserDisplayName, &dep.Status, &dep.CreatedAt, &dep.UpdatedAt,
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

func (s *Store) GetDeployments(ctx context.Context, workspaceID, repoID string) ([]domain.AppDeployment, error) {
	query, args, err := s.sq.Select("d.id", "d.fromDeploymentId", "d.repoId", "d.space", "d.sha", "d.branch", "d.commitMessage", "d.buildTag", "d.userDisplayName", "d.status", "d.createdAt", "d.updatedAt").
		From("deployments d").
		Join("installedRepos r ON d.repoId = r.id").
		Where(sq.And{
			sq.Eq{"d.repoId": repoID},
			sq.Eq{"r.workspaceId": workspaceID},
		}).
		OrderBy("d.id DESC").
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
		if err := rows.Scan(&dep.ID, &dep.FromDeploymentID, &dep.RepoID, &spacePayload, &dep.Sha, &dep.Branch, &dep.CommitMessage, &dep.BuildTag, &dep.UserDisplayName, &dep.Status, &dep.CreatedAt, &dep.UpdatedAt); err != nil {
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

func (s *Store) DeploymentBelongsToWorkspace(ctx context.Context, workspaceID, deploymentID string) (bool, error) {
	query, args, err := s.sq.Select("1").
		From("deployments d").
		Join("installedRepos r ON d.repoId = r.id").
		Where(sq.And{
			sq.Eq{"d.id": deploymentID},
			sq.Eq{"r.workspaceId": workspaceID},
		}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build DeploymentBelongsToWorkspace query: %w", err)
	}

	var exists int
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to query DeploymentBelongsToWorkspace: %w", err)
	}
	return true, nil
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

	workspace, err := s.GetDefaultWorkspace(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			// User exists but doesn't have a default workspace - create one
			workspace, err = s.createDefaultWorkspaceForUser(ctx, tx, userID)
			if err != nil {
				return "", fmt.Errorf("failed to create default workspace for user: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to get default workspace: %w", err)
		}
	}

	treenqInstallationID, err := s.linkGithub(ctx, tx, installationID, workspace.ID, repos)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit LinkGithub: %w", err)
	}
	return treenqInstallationID, nil
}

func (s *Store) linkGithub(ctx context.Context, querier Querier, installationID int, workspaceID string, repos []domain.InstalledRepository) (string, error) {
	timestamp := now()

	// save a github app
	treenqInstallationID, err := s.saveInstallation(ctx, querier, workspaceID, installationID, timestamp)
	if err != nil {
		return "", fmt.Errorf("failed to save installation: %w", err)
	}

	// clean previous state
	err = s.removeInstalledRepos(ctx, workspaceID, installationID, querier)
	if err != nil {
		return "", fmt.Errorf("failed to remove installed repos: %w", err)
	}

	// Insert repositories
	err = s.insertInstalledRepos(ctx, repos, workspaceID, installationID, timestamp, querier)
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
		_, err := s.linkGithub(ctx, tx, installationID, profile.CurrentWorkspace, repos)
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
	workspaceID string,
	installationID int,
	q Querier,
) error {
	repoQuery := s.sq.Delete("installedRepos").Where(sq.Eq{
		"workspaceId":    workspaceID,
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
	workspaceID string,
	installationID int,
	createdAt time.Time,
	q Querier,
) error {
	if len(repos) == 0 {
		return nil
	}

	repoQuery := s.sq.Insert("installedRepos").
		Columns("id", "githubId", "fullName", "private", "branch", "installationId", "workspaceId", "status", "createdAt")

	for _, repo := range repos {
		id := xid.New().String()
		repoQuery = repoQuery.Values(
			id,
			repo.ID,
			repo.FullName,
			repo.Private,
			"",
			installationID,
			workspaceID,
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	userID, err := s.getUserIDByDisplayName(ctx, senderLogin, s.db)
	if err != nil {
		return err
	}

	workspace, err := s.GetDefaultWorkspace(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			// User exists but doesn't have a default workspace - create one
			workspace, err = s.createDefaultWorkspaceForUser(ctx, tx, userID)
			if err != nil {
				return fmt.Errorf("failed to create default workspace for user: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get default workspace: %w", err)
		}
	}

	timestamp := now()
	err = s.insertInstalledRepos(ctx, repos, workspace.ID, installationID, timestamp, tx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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

func (s *Store) GetInstallationID(ctx context.Context, workspaceID, repoName string) (int, error) {
	q, args, err := s.sq.Select("installationId").
		From("installedRepos").
		Where(sq.Eq{"workspaceId": workspaceID, "fullName": repoName}).
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

func (s *Store) saveInstallation(ctx context.Context, q Querier, workspaceID string, githubID int, timestamp time.Time) (string, error) {
	// First check if installation already exists
	query, args, err := s.sq.Select("id").
		From("installations").
		Where(sq.Eq{
			"githubId":    githubID,
			"workspaceId": workspaceID,
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
		Columns("id", "workspaceId", "githubId", "createdAt").
		Values(installationID, workspaceID, githubID, timestamp).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build save installation query: %w", err)
	}

	if _, err := q.ExecContext(ctx, query, args...); err != nil {
		return "", fmt.Errorf("failed to save installation: %w", err)
	}

	return installationID, nil
}

func (s *Store) GetGithubRepos(ctx context.Context, workspaceID string) ([]domain.GithubRepository, bool, error) {
	hasInstallation, err := s.workspaceHasInstallation(ctx, workspaceID)
	if err != nil {
		return nil, false, nil
	}
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "status", "branch").
		From("installedRepos").
		Where(sq.Eq{"workspaceId": workspaceID}).
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

	// put connected repositories on top
	slices.SortFunc(repos, func(a, b domain.GithubRepository) int {
		if a.Branch == b.Branch {
			return 0
		}
		if a.Branch != "" {
			return -1
		}
		return 1
	})

	return repos, hasInstallation, nil
}

func (s *Store) workspaceHasInstallation(ctx context.Context, workspaceID string) (bool, error) {
	query, args, err := s.sq.Select("count(*)").
		From("installations").
		Where(sq.Eq{"workspaceId": workspaceID}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build workspaceHasInstallation query: %w", err)
	}

	var count int
	row := s.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to scan workspaceHasInstallation: %w", err)
	}

	return count > 0, nil
}

func (s *Store) ConnectRepo(ctx context.Context, workspaceID, repoID, branch string, space tqsdk.Space) (domain.GithubRepository, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.GithubRepository{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	repo, err := s.connectRepo(ctx, tx, workspaceID, repoID, branch)
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

func (s *Store) connectRepo(ctx context.Context, q Querier, workspaceID, repoID, branch string) (domain.GithubRepository, error) {
	query, args, err := s.sq.Update("installedRepos").
		Set("branch", branch).
		Where(sq.Eq{"id": repoID, "workspaceId": workspaceID}).
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

func (s *Store) GetRepoByID(ctx context.Context, workspaceID string, repoID string) (domain.GithubRepository, error) {
	var repo domain.GithubRepository
	query, args, err := s.sq.Select("id", "githubId", "fullName", "private", "branch", "installationId", "status").
		From("installedRepos").
		Where(sq.Eq{"id": repoID, "workspaceId": workspaceID}).
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

func (s *Store) SaveSecret(ctx context.Context, repoID, key, workspaceID string) error {
	createdAt := now()
	query, args, err := s.sq.Insert("secrets").
		Columns("repoId", "key", "workspaceId", "createdAt").
		Values(repoID, key, workspaceID, createdAt).
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

func (s *Store) GetRepositorySecretKeys(ctx context.Context, repoID, workspaceID string) ([]string, error) {
	query, args, err := s.sq.Select("key").
		From("secrets").
		Where(sq.Eq{"repoId": repoID, "workspaceId": workspaceID}).
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

func (s *Store) RepositorySecretKeyExists(ctx context.Context, repoID, key, workspaceID string) (bool, error) {
	query, args, err := s.sq.Select("1").
		From("secrets").
		Where(sq.Eq{"repoId": repoID, "workspaceId": workspaceID, "key": key}).
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

func (s *Store) RemoveSecret(ctx context.Context, repoID, key, workspaceID string) error {
	query, args, err := s.sq.Delete("secrets").
		Where(sq.Eq{"repoId": repoID, "key": key, "workspaceId": workspaceID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build RemoveSecret query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec RemoveSecret: %w", err)
	}

	return nil
}

func (s *Store) RemoveInstallation(ctx context.Context, installationID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start remove installation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// First get all repos for this installation
	repoQuery, args, err := s.sq.Select("id").
		From("installedRepos").
		Where(sq.Eq{"installationId": installationID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build repo query: %w", err)
	}

	rows, err := tx.QueryContext(ctx, repoQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to query repos: %w", err)
	}
	defer rows.Close()

	var repoIDs []string
	for rows.Next() {
		var repoID string
		if err := rows.Scan(&repoID); err != nil {
			return fmt.Errorf("failed to scan repo ID: %w", err)
		}
		repoIDs = append(repoIDs, repoID)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate repos: %w", err)
	}

	// Delete all related data for each repo
	for _, repoID := range repoIDs {
		// Delete secrets
		if _, err := s.sq.Delete("secrets").
			Where(sq.Eq{"repoId": repoID}).
			RunWith(tx).
			ExecContext(ctx); err != nil {
			return fmt.Errorf("failed to delete secrets for repo %s: %w", repoID, err)
		}

		// Delete spaces
		if _, err := s.sq.Delete("spaces").
			Where(sq.Eq{"repoId": repoID}).
			RunWith(tx).
			ExecContext(ctx); err != nil {
			return fmt.Errorf("failed to delete spaces for repo %s: %w", repoID, err)
		}

		// Delete deployments
		if _, err := s.sq.Delete("deployments").
			Where(sq.Eq{"repoId": repoID}).
			RunWith(tx).
			ExecContext(ctx); err != nil {
			return fmt.Errorf("failed to delete deployments for repo %s: %w", repoID, err)
		}
	}

	// Delete all repos for this installation
	if _, err := s.sq.Delete("installedRepos").
		Where(sq.Eq{"installationId": installationID}).
		RunWith(tx).
		ExecContext(ctx); err != nil {
		return fmt.Errorf("failed to delete repos: %w", err)
	}

	// Finally delete the installation
	if _, err := s.sq.Delete("installations").
		Where(sq.Eq{"githubId": installationID}).
		RunWith(tx).
		ExecContext(ctx); err != nil {
		return fmt.Errorf("failed to delete installation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit RemoveInstallation: %w", err)
	}

	return nil
}

func (s *Store) GetUserWorkspaces(ctx context.Context, userID string) ([]domain.Workspace, error) {
	query, args, err := s.sq.Select("w.id", "w.name", "w.githubOrgName", "wu.role").
		From("workspaces w").
		Join("workspaceUsers wu ON w.id = wu.workspaceId").
		Where(sq.Eq{"wu.userId": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build GetUserWorkspaces query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetUserWorkspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []domain.Workspace
	var personal domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		var githubOrgName string
		if err := rows.Scan(&workspace.ID, &workspace.Name, &githubOrgName, &workspace.Role); err != nil {
			return nil, fmt.Errorf("failed to scan GetUserWorkspaces row: %w", err)
		}
		if githubOrgName != "" {
			workspace.GithubOrgName = githubOrgName
		}
		// default workspace is personal and ever goes first
		if workspace.Name == personalWorkspaceName {
			personal = workspace
		} else {
			workspaces = append(workspaces, workspace)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate GetUserWorkspaces rows: %w", err)
	}

	return append([]domain.Workspace{personal}, workspaces...), nil
}

func (s *Store) GetDefaultWorkspace(ctx context.Context, userID string) (domain.Workspace, error) {
	query, args, err := s.sq.Select("w.id", "w.name", "w.githubOrgName", "wu.role").
		From("workspaces w").
		Join("workspaceUsers wu ON w.id = wu.workspaceId").
		Where(sq.Eq{"wu.userId": userID, "w.name": personalWorkspaceName}).
		ToSql()
	if err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to build GetDefaultWorkspace query: %w", err)
	}

	var workspace domain.Workspace
	var githubOrgName string
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&workspace.ID, &workspace.Name, &githubOrgName, &workspace.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return workspace, domain.ErrWorkspaceNotFound
		}
		return workspace, fmt.Errorf("failed to scan GetDefaultWorkspace: %w", err)
	}

	if githubOrgName != "" {
		workspace.GithubOrgName = githubOrgName
	}

	return workspace, nil
}

func (s *Store) GetWorkspaceByID(ctx context.Context, workspaceID string) (domain.Workspace, error) {
	query, args, err := s.sq.Select("id", "name", "githubOrgName").
		From("workspaces").
		Where(sq.Eq{"id": workspaceID}).
		ToSql()
	if err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to build GetWorkspaceByID query: %w", err)
	}

	var workspace domain.Workspace
	var githubOrgName string
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&workspace.ID, &workspace.Name, &githubOrgName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return workspace, domain.ErrWorkspaceNotFound
		}
		return workspace, fmt.Errorf("failed to scan GetWorkspaceByID: %w", err)
	}

	if githubOrgName != "" {
		workspace.GithubOrgName = githubOrgName
	}

	return workspace, nil
}

func (s *Store) GetWorkspaceByUserDisplayName(ctx context.Context, userDisplayName string) (domain.Workspace, error) {
	query, args, err := s.sq.Select("w.id", "w.name", "w.githubOrgName", "wu.role").
		From("workspaces w").
		Join("workspaceUsers wu ON w.id = wu.workspaceId").
		Join("users u ON wu.userId = u.id").
		Where(sq.Eq{"u.displayName": userDisplayName}).
		Limit(1).
		ToSql()
	if err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to build GetWorkspaceByUserDisplayName query: %w", err)
	}

	var workspace domain.Workspace
	var githubOrgName string
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&workspace.ID, &workspace.Name, &githubOrgName, &workspace.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return workspace, domain.ErrWorkspaceNotFound
		}
		return workspace, fmt.Errorf("failed to scan GetWorkspaceByUserDisplayName: %w", err)
	}

	if githubOrgName != "" {
		workspace.GithubOrgName = githubOrgName
	}

	return workspace, nil
}

// createDefaultWorkspaceForUser creates a default workspace for a user within an existing transaction
func (s *Store) createDefaultWorkspaceForUser(ctx context.Context, tx *sql.Tx, userID string) (domain.Workspace, error) {
	workspaceID := xid.New().String()
	workspaceName := xid.New().String()

	// Create the workspace
	workspaceQuery, workspaceArgs, err := s.sq.Insert("workspaces").
		Columns("id", "name").
		Values(workspaceID, personalWorkspaceName).
		ToSql()
	if err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to build workspace query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, workspaceQuery, workspaceArgs...); err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to create default workspace: %w", err)
	}

	// Link user to workspace
	userWorkspaceQuery, userWorkspaceArgs, err := s.sq.Insert("workspaceUsers").
		Columns("workspaceId", "userId", "role").
		Values(workspaceID, userID, "admin").
		ToSql()
	if err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to build workspace user query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, userWorkspaceQuery, userWorkspaceArgs...); err != nil {
		return domain.Workspace{}, fmt.Errorf("failed to add user to workspace: %w", err)
	}

	return domain.Workspace{
		ID:   workspaceID,
		Name: workspaceName,
		Role: "admin",
	}, nil
}
