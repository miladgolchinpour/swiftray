package models

import "github.com/google/uuid"

type Subscription struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	Nodes       []Node  `json:"nodes"`
	LastUpdated *string `json:"lastUpdated"`
}

func NewSubscription(name, url string) Subscription {
	return Subscription{
		ID:   uuid.New().String(),
		Name: name,
		URL:  url,
		Nodes: []Node{},
	}
}
