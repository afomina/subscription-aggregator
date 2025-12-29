package repository

import (
	"database/sql"
	"subscription-aggregator/internal/model"
	"github.com/google/uuid"
)

type SubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) scanRow(scanner interface{ Scan(...interface{}) error }) (*model.Subscription, error) {
	sub := &model.Subscription{}
	var endDate sql.NullTime
	err := scanner.Scan(&sub.ID, &sub.ServiceName, &sub.CostRub, &sub.UserID, &sub.StartDate, &endDate)
	if err != nil {
		return nil, err
	}
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	return sub, nil
}

func (r *SubscriptionRepo) Create(sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, cost_rub, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.db.QueryRow(query, sub.ServiceName, sub.CostRub, sub.UserID, sub.StartDate, sub.EndDate).Scan(&sub.ID)
}

func (r *SubscriptionRepo) GetByID(id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, service_name, cost_rub, user_id, start_date, end_date FROM subscriptions WHERE id = $1`
	row := r.db.QueryRow(query, id)
	return r.scanRow(row)
}

func (r *SubscriptionRepo) GetAll() ([]*model.Subscription, error) {
	rows, err := r.db.Query(`SELECT id, service_name, cost_rub, user_id, start_date, end_date FROM subscriptions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []*model.Subscription
	for rows.Next() {
		sub, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *SubscriptionRepo) Update(id uuid.UUID, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions 
		SET service_name = $1, cost_rub = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6`
	_, err := r.db.Exec(query, sub.ServiceName, sub.CostRub, sub.UserID, sub.StartDate, sub.EndDate, id)
	return err
}

func (r *SubscriptionRepo) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM subscriptions WHERE id = $1`, id)
	return err
}

func (r *SubscriptionRepo) GetTotalCost(
	userID *uuid.UUID,
	serviceName *string,
	startMonth, startYear, endMonth, endYear int,
) (int, error) {
	query := `
		WITH period_months AS (
			SELECT generate_series(
				DATE_TRUNC('month', make_date($3, $4, 1)::timestamp),
				DATE_TRUNC('month', make_date($5, $6, 1)::timestamp),
				'1 month'
			)::date AS month_start
		)
		SELECT COALESCE(SUM(s.cost_rub), 0)
		FROM subscriptions s
		JOIN period_months pm ON
			pm.month_start >= DATE_TRUNC('month', s.start_date) AND
			(s.end_date IS NULL OR pm.month_start <= DATE_TRUNC('month', COALESCE(s.end_date, '9999-12-31'::date)))
		WHERE
			($1::uuid IS NULL OR s.user_id = $1::uuid) AND
			($2::text IS NULL OR s.service_name = $2::text)
	`

	var total int
	err := r.db.QueryRow(query,
		userID,
		serviceName,
		startYear,
		startMonth,
		endYear,
		endMonth,
	).Scan(&total)

	return total, err
}