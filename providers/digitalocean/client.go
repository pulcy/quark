package digitalocean

import (
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func NewDOClient(token string) *godo.Client {
	tokenSource := &TokenSource{
		AccessToken: token,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	return client
}

func DropletList(client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(opt)
		if err != nil {
			return list, err
		}

		// append the current page's droplets to our list
		for _, d := range droplets {
			list = append(list, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return list, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

func ImageList(client *godo.Client) ([]godo.Image, error) {
	// create a list to hold our images
	list := []godo.Image{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		images, resp, err := client.Images.List(opt)
		if err != nil {
			return list, err
		}

		// append the current page's images to our list
		list = append(list, images...)

		// if we are at the last page, break out the for loop
		if resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return list, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}
func KeyList(client *godo.Client) ([]godo.Key, error) {
	// create a list to hold our keys
	list := []godo.Key{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		keys, resp, err := client.Keys.List(opt)
		if err != nil {
			return list, err
		}

		// append the current page's keys to our list
		list = append(list, keys...)

		// if we are at the last page, break out the for loop
		if resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return list, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

func RegionList(client *godo.Client) ([]godo.Region, error) {
	// create a list to hold our regions
	list := []godo.Region{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		regions, resp, err := client.Regions.List(opt)
		if err != nil {
			return list, err
		}

		// append the current page's droplets to our list
		list = append(list, regions...)

		// if we are at the last page, break out the for loop
		if resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return list, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

func DomainRecordList(client *godo.Client, domain string) ([]godo.DomainRecord, error) {
	// create a list to hold our records
	list := []godo.DomainRecord{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		records, resp, err := client.Domains.Records(domain, opt)
		if err != nil {
			return list, err
		}

		// append the current page's records to our list
		list = append(list, records...)

		// if we are at the last page, break out the for loop
		if resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return list, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}
