package service

import (
	"subscription-aggregator/internal/model"
	"subscription-aggregator/internal/repository"
	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepo
}

func NewSubscriptionService(repo *repository.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(sub *model.Subscription) error {
	return s.repo.Create(sub)
}

func (s *SubscriptionService) GetByID(id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *SubscriptionService) GetAll() ([]*model.Subscription, error) {
	return s.repo.GetAll()
}

func (s *SubscriptionService) Update(id uuid.UUID, sub *model.Subscription) error {
	return s.repo.Update(id, sub)
}

func (s *SubscriptionService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *SubscriptionService) GetTotalCost(
	userID *uuid.UUID,
	serviceName *string,
	startMonth, startYear, endMonth, endYear int,
) (int, error) {
	return s.repo.GetTotalCost(userID, serviceName, startMonth, startYear, endMonth, endYear)
}