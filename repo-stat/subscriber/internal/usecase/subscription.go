package usecase

import (
	"context"
	"errors"
	"strings"

	db "repo-stat/subscriber/internal/sqlc"

	"github.com/jackc/pgx/v5/pgconn"
)

type SubscriptionStore interface {
	CreateSubscription(ctx context.Context, arg db.CreateSubscriptionParams) (db.Subscription, error)
	DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) (int64, error)
	ListSubscriptions(ctx context.Context) ([]db.Subscription, error)
}

type Subscription struct {
	store SubscriptionStore
}

type SubscriptionModel struct {
	Owner    string
	RepoName string
}

func NewSubscription(store SubscriptionStore) *Subscription {
	return &Subscription{store: store}
}

func (u *Subscription) Create(ctx context.Context, owner, repo string) (*SubscriptionModel, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" {
		return nil, ErrOwnerRequired
	}
	if repo == "" {
		return nil, ErrRepoNameRequired
	}

	subscription, err := u.store.CreateSubscription(ctx, db.CreateSubscriptionParams{
		RepoOwner: owner,
		RepoName:  repo,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return &SubscriptionModel{
		Owner:    subscription.RepoOwner,
		RepoName: subscription.RepoName,
	}, nil
}

func (u *Subscription) Delete(ctx context.Context, owner, repo string) error {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" {
		return ErrOwnerRequired
	}
	if repo == "" {
		return ErrRepoNameRequired
	}

	rowsAffected, err := u.store.DeleteSubscription(ctx, db.DeleteSubscriptionParams{
		RepoOwner: owner,
		RepoName:  repo,
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRepoNotFound
	}

	return nil
}

func (u *Subscription) List(ctx context.Context) ([]SubscriptionModel, error) {
	subscriptions, err := u.store.ListSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]SubscriptionModel, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, SubscriptionModel{
			Owner:    subscription.RepoOwner,
			RepoName: subscription.RepoName,
		})
	}

	return result, nil
}
