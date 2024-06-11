package handlers

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
)

type Do_client struct {
	client *godo.Client
}

type Config struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Size   string `json:"size"`
	Image  string `json:"image"`
}

type TokenRequest struct {
	token string
}

type CreateDropletRequest struct {
	Name string
}

type DropletListResponse struct {
	DropletList interface{} `json:"DropletList"`
}

func (d *Do_client) authClient(ctx context.Context, req TokenRequest) (Config, *Error) {
	client, err := godo.NewFromToken(os.Getenv("DIGITAL_OCEAN_TOKEN"))
	if err != nil {
		return Config{}, err.Error()
	}

	return CreateDropletRequest(client)
}

func (d *Do_client) CreateDroplet(ctx context.Context, client *godo.Client, req CreateDropletRequest) (Config, *Error) {
	createRequest := &godo.DropletCreateRequest{
		Name:   req.Name,
		Region: "nyc3",
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-20-04-x64",
		},
	}

	_, _, err := client.Droplets.Create(ctx, createRequest)

	if err != nil {
		return Config{}, err.Error()
	}

	return Config{
		Name:   req.Name,
		Region: "nyc3",
		Size:   "s-1vcpu-1gb",
		Image:  "ubuntu-20-04-x64",
	}, nil
}

func (d *Do_client) DropletList(ctx context.Context, client *godo.Client) (DropletListResponse, Error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}

	droplets, _, err := client.Droplets.List(ctx, opt)
	if err != nil {
		return DropletListResponse{}, err
	}

	// append the current page's droplets to our list
	list = append(list, droplets...)

	return list, Error{}
}

func (d *Do_client) GetDroplet(ctx context.Context, client *godo.Client, id int) (*godo.Droplet, Error) {
	//getdroplet by id
	droplet, _, err := client.Droplets.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return droplet, Error{}
}
