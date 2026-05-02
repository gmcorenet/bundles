package service

import (
	"github.com/gmcorenet/skeleton/internal/model"
	"github.com/gmcorenet/skeleton/internal/repository"
)

type Service interface {
	GetAll() []model.Entity
	GetByID(id string) *model.Entity
	Create(entity *model.Entity) *model.Entity
	Update(id string, entity *model.Entity) *model.Entity
	Delete(id string) bool
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll() []model.Entity {
	return s.repo.FindAll()
}

func (s *service) GetByID(id string) *model.Entity {
	return s.repo.FindByID(id)
}

func (s *service) Create(entity *model.Entity) *model.Entity {
	return s.repo.Create(entity)
}

func (s *service) Update(id string, entity *model.Entity) *model.Entity {
	return s.repo.Update(id, entity)
}

func (s *service) Delete(id string) bool {
	return s.repo.Delete(id)
}