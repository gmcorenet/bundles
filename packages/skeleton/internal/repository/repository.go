package repository

import "github.com/gmcorenet/skeleton/internal/model"

type Repository interface {
	FindAll() []model.Entity
	FindByID(id string) *model.Entity
	Create(entity *model.Entity) *model.Entity
	Update(id string, entity *model.Entity) *model.Entity
	Delete(id string) bool
}

type baseRepository struct {
	items []model.Entity
}

func NewRepository() Repository {
	return &baseRepository{
		items: []model.Entity{},
	}
}

func (r *baseRepository) FindAll() []model.Entity {
	return r.items
}

func (r *baseRepository) FindByID(id string) *model.Entity {
	for _, item := range r.items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

func (r *baseRepository) Create(entity *model.Entity) *model.Entity {
	r.items = append(r.items, *entity)
	return entity
}

func (r *baseRepository) Update(id string, entity *model.Entity) *model.Entity {
	for i, item := range r.items {
		if item.ID == id {
			r.items[i] = *entity
			return entity
		}
	}
	return nil
}

func (r *baseRepository) Delete(id string) bool {
	for i, item := range r.items {
		if item.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return true
		}
	}
	return false
}