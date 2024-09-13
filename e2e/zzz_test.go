package e2e

import (
	"net/http"

	"github.com/treenq/treenq/client"
)

var config = struct {
	AuthToken string
}{
	AuthToken: "vWykYUlys5bxEFY0CV8v1NbSyKtaIStBEtnoZlgcBtJDN-cZcnKnxc1xcvoF3JciAPYzj5I",
}

var apiClient = client.NewClient("http://localhost:8000", http.DefaultClient, map[string]string{
	"Authorization": "Bearer " + config.AuthToken,
})
