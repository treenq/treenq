package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/treenq/treenq/src/domain"
)

var ErrProviderNotFound = errors.New("provider not found")

type Zitadel struct {
	client *http.Client

	domain     string
	token      string
	idps       map[string]string
	successUrl string
	failureUrl string
}

func NewZitadel(
	domain string,
	token string,
	idps map[string]string,
	successUrl string,
	failureUrl string,
) *Zitadel {
	return &Zitadel{
		client: http.DefaultClient,

		domain:     domain,
		token:      token,
		idps:       idps,
		successUrl: successUrl,
		failureUrl: failureUrl,
	}
}

type intentRequest struct {
	IDPId string     `json:"idpId"`
	Urls  intentUrls `json:"urls"`
}

type intentUrls struct {
	SuccessURL string `json:"successUrl"`
	FailureURL string `json:"failureUrl"`
}

type intentResponse struct {
	AuthURL string `json:"authUrl"`
}

func (z *Zitadel) Start(ctx context.Context, provider string) (string, error) {
	idpID, ok := z.idps[provider]
	if !ok {
		return "", ErrProviderNotFound
	}

	// Prepare the request payload
	reqPayload := intentRequest{
		IDPId: idpID,
		Urls: intentUrls{
			SuccessURL: z.successUrl,
			FailureURL: z.failureUrl,
		},
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	url := fmt.Sprintf("%s/v2/idp_intents", z.domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+z.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected response: %s", string(body))
	}

	var respPayload intentResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respPayload.AuthURL, nil
}

type idpIntentRequest struct {
	IdpIntentToken string `json:"idpIntentToken"`
}

type UserResponse struct {
	Details        UserDetails `json:"details"`
	IdpInformation IdpInfo     `json:"idpInformation"`
	UserID         string      `json:"userId"`
}

type UserDetails struct {
	ChangeDate    string `json:"changeDate"`
	ResourceOwner string `json:"resourceOwner"`
	Sequence      string `json:"sequence"`
}

type IdpInfo struct {
	IdpID          string `json:"idpId"`
	Oauth          Oauth  `json:"oauth"`
	RawInformation Raw    `json:"rawInformation"`
	UserID         string `json:"userId"`
	UserName       string `json:"userName"`
}

type Oauth struct {
	AccessToken string `json:"accessToken"`
}

type Raw struct {
	AvatarURL               string `json:"avatar_url"`
	Bio                     string `json:"bio"`
	Blog                    string `json:"blog"`
	Collaborators           int64  `json:"collaborators"`
	Company                 string `json:"company"`
	CreatedAt               string `json:"created_at"`
	DiskUsage               int64  `json:"disk_usage"`
	Email                   string `json:"email"`
	EventsURL               string `json:"events_url"`
	Followers               int64  `json:"followers"`
	FollowersURL            string `json:"followers_url"`
	Following               int64  `json:"following"`
	FollowingURL            string `json:"following_url"`
	GistsURL                string `json:"gists_url"`
	GravatarID              string `json:"gravatar_id"`
	Hireable                bool   `json:"hireable"`
	HTMLURL                 string `json:"html_url"`
	ID                      int64  `json:"id"`
	Location                string `json:"location"`
	Login                   string `json:"login"`
	Name                    string `json:"name"`
	NodeID                  string `json:"node_id"`
	OrganizationsURL        string `json:"organizations_url"`
	OwnedPrivateRepos       int64  `json:"owned_private_repos"`
	Plan                    Plan   `json:"plan"`
	PrivateGists            int64  `json:"private_gists"`
	PublicGists             int64  `json:"public_gists"`
	PublicRepos             int64  `json:"public_repos"`
	ReceivedEventsURL       string `json:"received_events_url"`
	ReposURL                string `json:"repos_url"`
	SiteAdmin               bool   `json:"site_admin"`
	StarredURL              string `json:"starred_url"`
	SubscriptionsURL        string `json:"subscriptions_url"`
	TotalPrivateRepos       int64  `json:"total_private_repos"`
	TwitterUsername         string `json:"twitter_username"`
	TwoFactorAuthentication bool   `json:"two_factor_authentication"`
	Type                    string `json:"type"`
	UpdatedAt               string `json:"updated_at"`
	URL                     string `json:"url"`
}

type Plan struct {
	Collaborators int64  `json:"collaborators"`
	Name          string `json:"name"`
	PrivateRepos  int64  `json:"private_repos"`
	Space         int64  `json:"space"`
}

func (z *Zitadel) GetIdpUser(ctx context.Context, intent, token string) (domain.UserInfo, error) {
	// Prepare the request payload
	reqPayload := idpIntentRequest{
		IdpIntentToken: token,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	url := fmt.Sprintf("%s/v2/idp_intents/%s", z.domain, intent)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+z.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return domain.UserInfo{}, fmt.Errorf("unexpected response: %s", string(body))
	}

	var respPayload UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return domain.UserInfo{
		Email:       respPayload.IdpInformation.RawInformation.Email,
		DisplayName: respPayload.IdpInformation.UserName,
	}, nil
}

type UserHuman struct {
	Username string      `json:"username"`
	Profile  UserProfile `json:"profile"`
	Email    UserEmail   `json:"email"`
	IdpLinks []IdpLink   `json:"idpLinks"`
}

type UserProfile struct {
	GivenName         string `json:"givenName"`
	FamilyName        string `json:"familyName"`
	DisplayName       string `json:"displayName"`
	PreferredLanguage string `json:"preferredLanguage"`
	NickName          string `json:"nickName"`
	Gender            string `json:"gender"`
}

type UserEmail struct {
	Email      string `json:"email"`
	IsVerified bool   `json:"isVerified"`
}

type IdpLink struct {
	IdpId    string `json:"idpId"`
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
}

type CreateUserResponse struct {
	UserId string `json:"userId"`
}

func (z *Zitadel) CreateUser(ctx context.Context, r domain.UserInfo) (domain.UserInfo, error) {
	reqPayload := UserHuman{
		Username: r.Email,
		Profile: UserProfile{
			GivenName:   "-",
			FamilyName:  "-",
			DisplayName: r.DisplayName,
		},
		Email: UserEmail{
			Email:      r.Email,
			IsVerified: true,
		},
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	url := fmt.Sprintf("%s/v2/users/human", z.domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+z.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return domain.UserInfo{}, fmt.Errorf("unexpected response: %s", string(body))
	}

	var res CreateUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return domain.UserInfo{}, fmt.Errorf("faeild to decode: %w", err)
	}

	return domain.UserInfo{ID: res.UserId, Email: r.Email, DisplayName: r.DisplayName}, nil
}

type ListUsersRequest struct {
	Query         ListUserPageQuery `json:"query"`
	SortingColumn string            `json:"sortingColumn"`
	Queries       []ListUserQuery   `json:"queries"`
}

type ListUserPageQuery struct {
	Offset int  `json:"offset"`
	Limit  int  `json:"limit"`
	Asc    bool `json:"asc"`
}

type ListUserQuery struct {
	EmailQuery ListUserEmailQuery `json:"emailQuery,omitempty"`
}

type ListUserEmailQuery struct {
	EmailAddress string `json:"emailAddress"`
	Method       string `json:"method"`
}

type ListUsersResponse struct {
	Result []ListUserResult `json:"result"`
}

type ListUserResult struct {
	UserID string    `json:"userId"`
	Human  UserHuman `json:"human"`
}

func (z *Zitadel) GetUserByEmail(ctx context.Context, email string) (domain.UserInfo, error) {
	reqPayload := ListUsersRequest{
		Query: ListUserPageQuery{
			Offset: 0,
			Limit:  1,
			Asc:    true,
		},
		SortingColumn: "USER_FIELD_NAME_UNSPECIFIED",
		Queries: []ListUserQuery{
			{
				EmailQuery: ListUserEmailQuery{
					EmailAddress: email,
					Method:       "TEXT_QUERY_METHOD_EQUALS",
				},
			},
		},
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	url := fmt.Sprintf("%s/v2/users", z.domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+z.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return domain.UserInfo{}, fmt.Errorf("unexpected response: %s", string(body))
	}

	var res ListUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(res.Result) == 0 {
		return domain.UserInfo{}, domain.ErrUserNotFound
	}

	user := res.Result[0]
	return domain.UserInfo{
		Email:       user.Human.Email.Email,
		DisplayName: user.Human.Profile.DisplayName,
	}, nil
}
