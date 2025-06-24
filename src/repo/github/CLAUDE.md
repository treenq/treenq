# GitHub Integration Package Documentation

## Overview

The GitHub integration in Treenq provides secure access to GitHub repositories and installations through the GitHub App model. This package handles authentication, repository discovery, and installation management.

## Architecture

### Package Structure

```
src/repo/github.go              # GitHub API client implementation
src/services/auth/github.go     # OAuth provider with token caching
src/domain/syncGithubApp.go     # Sync handler for refreshing installations
```

### Key Components

#### GitHub Client (`src/repo/github.go`)

- **Primary Purpose**: Interface with GitHub API using App installation tokens
- **Authentication**: Uses GitHub App JWT tokens for API access
- **Caching**: Implements token and branch caching for performance

**Key Methods:**

- `IssueAccessToken(installationID)` - Gets installation access tokens
- `GetUserInstallation(displayName)` - Gets user's personal installation
- `GetUserAccessibleInstallations(userToken)` - Lists installations accessible to user token
- `ListAllRepositoriesForUser(userToken)` - Fetches repos from all user-accessible installations
- `GetBranches(installationID, owner, repo)` - Gets repository branches with caching

#### OAuth Provider (`src/services/auth/github.go`)

- **Primary Purpose**: Handles GitHub OAuth flow and token management
- **Security Feature**: 15-minute token caching to avoid database storage
- **Package**: `auth` (not `github` to avoid circular imports)

**Key Methods:**

- `ExchangeUser(code)` - Exchanges OAuth code for user info and caches token
- `GetUserGithubToken(userDisplayName)` - Retrieves cached token or returns `ErrUnauthorized`

#### Sync Handler (`src/domain/syncGithubApp.go`)

- **Primary Purpose**: Refreshes all GitHub installations for a user
- **Use Cases**: Local development, webhook failures, manual sync
- **Security**: Uses user's GitHub access token via OAuth provider cache

## GitHub API Endpoints Used

### App Authentication (JWT-based)

- `POST /app/installations/{id}/access_tokens` - Get installation access token
- `GET /users/{username}/installation` - Get user's installation
- `GET /installation/repositories` - List repositories for installation

### User Authentication (OAuth token-based)

- `GET /user/installations` - List installations accessible to user
- `GET /user` - Get user profile information
- `GET /user/emails` - Get user email addresses

## Security Model

### Token Management Evolution

- **OAuth Provider Caching**: 15-minute in-memory token cache
- **Database-free**: No GitHub tokens stored in persistent storage
- **Error Handling**: Returns `UNAUTHORIZED` for expired tokens
- **Re-authentication**: Forces users to re-authenticate when tokens expire

### Authentication Flows

#### GitHub App Installation Access

1. Generate GitHub App JWT token
2. Exchange for installation access token
3. Cache installation token with expiration
4. Use for repository operations

#### User Installation Discovery

1. User authenticates via OAuth
2. GitHub access token cached for 15 minutes
3. Use `/user/installations` endpoint with user token
4. Discover all accessible installations (personal + organizations)

## Implementation Details

#### 1. Token Caching Implementation

```go
// Added to GithubOauthProvider
tokenCache *cache.Cache[string, string] // userDisplayName -> token

// In ExchangeUser method
p.tokenCache.Set(userInfo.DisplayName, token.AccessToken, cache.WithExpiration(15*time.Minute))
```

#### 2. Error Handling

```go
// Domain error definition
var ErrUnauthorized = errors.New("unauthorized: github token expired or invalid")

// OAuth provider usage
func (p *GithubOauthProvider) GetUserGithubToken(userDisplayName string) (string, error) {
    token, ok := p.tokenCache.Get(userDisplayName)
    if !ok {
        return "", domain.ErrUnauthorized
    }
    return token, nil
}
```

#### 3. Sync Handler Updates

- Removed dependency on database-stored tokens
- Uses OAuth provider for token retrieval
- Proper error codes for unauthorized scenarios
- Maintains user context through JWT claims

### Security Considerations

#### Why User Tokens Are Required

- **Installation Discovery**: Only user tokens can access `/user/installations`
- **Scoped Access**: User tokens respect organization permissions
- **Security**: Prevents unauthorized access to installations

#### Token Caching Benefits

- **Performance**: Reduces GitHub API calls
- **Security**: Short-lived in-memory storage
- **Reliability**: Automatic expiration and cleanup
- **User Experience**: Seamless operation within 15-minute window

### Error Scenarios

#### Token Expiration

- **Trigger**: After 15 minutes of caching
- **Response**: `UNAUTHORIZED` error code
- **User Action**: Re-authenticate with GitHub
- **System Behavior**: Clears cache entry automatically

#### Installation Access Issues

- **Missing Installation**: Returns `ErrInstallationNotFound`
- **No Repositories**: Returns `NO_INSTALLATIONS_FOUND` error
- **API Errors**: Detailed error messages with HTTP status codes

## Usage Patterns

### Local Development

```go
// SyncGithubApp handler automatically:
// 1. Gets user token from OAuth cache
// 2. Discovers all accessible installations
// 3. Fetches repositories from each installation
// 4. Saves transactionally to database
```

### Production Webhooks

- GitHub sends webhook events for repository changes
- Automatic repository synchronization
- Installation management through webhook events

### Manual Sync

- User-triggered refresh of installations
- Handles webhook failures or missed events
- Comprehensive repository discovery

## Dependencies

### External Libraries

- `github.com/Code-Hex/go-generics-cache` - Generic in-memory caching
- `golang.org/x/oauth2` - OAuth2 client implementation

### Internal Dependencies

- `src/domain` - Business logic and error definitions
- `pkg/vel` - Web framework and HTTP utilities
- Repository interfaces for data persistence
