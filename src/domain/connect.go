package domain

type ConnectRequest struct {
	Url string `json:"url"`
}

type ConnectResponse struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}
