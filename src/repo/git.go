package repo

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"net/url"
	"os"
	"strings"
)

//Clone repo component

func main() {
	var repoURL string

	fmt.Printf("Enter Your GitHub Repo URL : ")
	fmt.Scan(&repoURL)

	repoName, err := getRepoId(repoURL)
	if err != nil {
		err := fmt.Errorf("%w",err)
		fmt.Println(err)
	}

	//intializing a temp directory

	tempfile, err := os.MkdirTemp("", repoName)
	if err != nil {
		err := fmt.Errorf("Failed to create TempDir !!!%w", err)
		fmt.Println(err)
	}

	defer os.Remove(tempfile) // to remove the file incase of error

	fmt.Println("Cloning the repo...")

	//using go-git lib to clone the repo
	_, err = git.PlainClone(tempfile, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	if err != nil {
		//return "", fmt.Errorf("Failed to clone the repo !!!%w", err)  
		//I'm getting error if I'm using the above line so for now I'm printing the err below like that
		err := fmt.Errorf("Failed to clone the repo !!!%w", err)
		fmt.Println(err)
	}

	fmt.Println("GitHub repo successfully cloned to", tempfile)

}

// extracting the repo name
func getRepoId(repoURL string) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	path := parsedURL.Path
	segments := strings.Split(path, "/")

	// to check if the URL is valid
	if len(segments) < 3 {
		return "", fmt.Errorf("Invalid GitHub repo URL %s : %w", repoURL, err)
	}

	repoName := segments[len(segments)-1]
	return repoName, nil
}
