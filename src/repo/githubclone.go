package repo

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

func  GitCloneTemp(repoID string,repo string,username string,accesstoken string)(string, error){
	
	dirName := fmt.Sprintf("%s-%s-%s",repoID,username,repo)
	dir,err := os.MkdirTemp("",dirName)
	if(err != nil){
		return "",fmt.Errorf("Error while creating Directory %s",err)
	}
	defer os.RemoveAll(dir)

	url := fmt.Sprintf("https://%s@github.com/%s/%s.git",accesstoken,username,repo)

	t, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
		Progress: os.Stdout,
	})
	if(err != nil){
		return "",fmt.Errorf("Error while cloning the repo %s",err)
	}
	// t should be used that is why printed as of now 
	fmt.Print(t)
  return dir,err
} 