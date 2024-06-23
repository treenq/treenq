package repo

import (
	"fmt"
	"net/http"
	"time"
	"os"
	"encoding/json"
	"github.com/go-git/go-git/v5"
)

type TokenIssuer interface {
	GeneratedJwtToken() (string, error)
}

type GithubClient struct {
	tokenIssuer TokenIssuer
	client *http.Client
}
func NewGithubClient(tokenIssuer TokenIssuer, client *http.Client) *GithubClient {
	if(client == nil){
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GithubClient{
	  tokenIssuer: tokenIssuer,
	  client: client,
	}
  }

var responseBody struct{
	Token string `json:"token"`
}

func (c *GithubClient)  IssueAccessToken(installationID string) (string,error) {
	jwtToken,err := c.tokenIssuer.GeneratedJwtToken()
	if(err != nil){
		return "",err
	}
	url :=fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens",installationID)
	req, err := http.NewRequest("POST", url, nil)
	if(err != nil){
		return "",fmt.Errorf("failed to create new request %s",err)
	}
	req.Header.Set("Authorization","Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return "", fmt.Errorf("An error occurred while processing request: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("Failed to decode response: %w", err)
	}

	return responseBody.Token,nil
}

func  GitCloneTemp(repoID string,repo string,username string,accesstoken string)(string, error){

	dirName := fmt.Sprintf("%s-%s-%s",repoID,username,repo)
	dir,err := os.MkdirTemp("",dirName)
	if(err != nil){
		return "",fmt.Errorf("Error while creating Directory %s",err)
	}
	url := fmt.Sprintf("https://%s@github.com/%s/%s.git",accesstoken,username,repo)
	_ , err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
		Progress: os.Stdout,
	})
	if(err != nil){
		return "",fmt.Errorf("Error while cloning the repo %s",err)
	}
  return dir,err
} 
