package repository

import (
	"encoding/json"
	"log"
	"net/url"
	"nokib/campwiz/models"
)

type Paginator[PageType any] struct {
	repo *CommonsRepository
}

func NewPaginator[PageType any](repo *CommonsRepository) *Paginator[PageType] {
	return &Paginator[PageType]{
		repo: repo,
	}
}

/*
This method would return a stream of pages from the commons API
It would handle the pagination of the API
*/
func (p *Paginator[PageType]) Query(params url.Values) (chan *PageType, error) {
	// Query
	streamChanel := make(chan *PageType)
	go func() {
		defer close(streamChanel)
		for {
			stream, err := p.repo.Get(params)
			if err != nil {
				log.Fatal(err)
			}
			defer stream.Close()
			resp := &PageQueryResponse[PageType]{}
			err = json.NewDecoder(stream).Decode(resp)
			if err != nil {
				log.Fatal(err)
			}
			for _, page := range resp.Query.Pages {
				streamChanel <- &page
			}
			Continue := resp.Next
			if Continue == nil {
				streamChanel <- nil
				break
			}
			// Convert to map
			for key, value := range *Continue {
				params.Set(key, value)
			}
		}
	}()
	return streamChanel, nil
}
func (p *Paginator[PageType]) UserList(params url.Values) (chan *models.WikimediaUser, error) {
	// Query
	streamChanel := make(chan *models.WikimediaUser)
	go func() {
		defer close(streamChanel)
		defer func() { streamChanel <- nil }()
		for {
			stream, err := p.repo.Get(params)
			if err != nil {
				log.Fatal(err)
			}
			defer stream.Close()
			resp := &UserListQueryResponse{}
			err = json.NewDecoder(stream).Decode(resp)
			if err != nil {
				log.Println(err)
				break
			}
			if resp.Error != nil {
				log.Println(resp.Error)
				break
			}
			for _, page := range resp.Query.Users {
				streamChanel <- &page
			}
			Continue := resp.Next
			if Continue == nil {
				break
			}
			// Convert to map
			for key, value := range *Continue {
				params.Set(key, value)
			}
		}
	}()
	return streamChanel, nil
}
