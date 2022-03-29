package service

import "github.com/ava-labs/avalanchego/database"

type Service struct {
	db database.Database
}

func (s *Service) Install(alias string) error {

}

func (s *Service) Uninstall(alias string) error {

}

func (s *Service) Update(alias string) error {

}

func (s *Service) Search(alias string) error {

}

func (s *Service) Info(alias string) error {

}

func (s *Service) Sync(repo string) error {

}

func (s *Service) AddRepository(repo string) error {

}

func (s *Service) RemoveRepository(repo string) error {

}

func (s *Service) ListRepositories() error {

}

func New() *Service {
	s := &Service{}

	return s
}
