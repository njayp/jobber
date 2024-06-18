package server

import "github.com/njayp/jobber/pkg/manager"

func NewService() (*Service, error) {
	m, err := manager.NewManager()
	if err != nil {
		return nil, err
	}
	return &Service{manager: m}, nil
}
