package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"subscription-aggregator/internal/model"
	"subscription-aggregator/internal/service"
)

type Handler struct {
	service *service.SubscriptionService
	logger  *logrus.Logger
}

func NewHandler(s *service.SubscriptionService, logger *logrus.Logger) *Handler {
	return &Handler{service: s, logger: logger}
}

// @Summary Create subscription
// @Description Create a new subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} map[string]interface{}
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	var req model.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid input")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		CostRub:     req.CostRub,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.service.Create(sub); err != nil {
		h.logger.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// @Summary Get subscription by ID
// @Description Get subscription by UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}
	sub, err := h.service.GetByID(id)
	if err != nil {
		h.logger.WithError(err).Error("Subscription not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// @Summary List all subscriptions
// @Description Get all subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {array} model.Subscription
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	subs, err := h.service.GetAll()
	if err != nil {
		h.logger.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, subs)
}

// @Summary Update subscription
// @Description Update subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body model.UpdateSubscriptionRequest true "Updated subscription data"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	var req model.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid update input")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ID берётся из URL, а не из тела запроса!
	sub := &model.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		CostRub:     req.CostRub,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.service.Update(id, sub); err != nil {
		h.logger.WithError(err).Error("Update failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// @Summary Delete subscription
// @Description Delete subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}
	if err := h.service.Delete(id); err != nil {
		h.logger.WithError(err).Error("Delete failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Get total cost
// @Description Calculate total subscription cost for period with optional filters
// @Tags analytics
// @Produce json
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Service name"
// @Param start_month query int true "Start month (1-12)"
// @Param start_year query int true "Start year"
// @Param end_month query int true "End month (1-12)"
// @Param end_year query int true "End year"
// @Success 200 {object} map[string]int
// @Router /subscriptions/total-cost [get]
func (h *Handler) GetTotalCost(c *gin.Context) {
	var userID *uuid.UUID
	if uidStr := c.Query("user_id"); uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		userID = &uid
	}

	serviceName := c.Query("service_name")
	if serviceName == "" {
		serviceName = ""
	}

	startMonth, _ := strconv.Atoi(c.Query("start_month"))
	startYear, _ := strconv.Atoi(c.Query("start_year"))
	endMonth, _ := strconv.Atoi(c.Query("end_month"))
	endYear, _ := strconv.Atoi(c.Query("end_year"))

	if startMonth < 1 || startMonth > 12 || endMonth < 1 || endMonth > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month must be 1-12"})
		return
	}

	if startYear > endYear || (startYear == endYear && startMonth > endMonth) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period"})
		return
	}

	var sn *string
	if serviceName != "" {
		sn = &serviceName
	}

	total, err := h.service.GetTotalCost(userID, sn, startMonth, startYear, endMonth, endYear)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate total cost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_cost_rub": total})
}