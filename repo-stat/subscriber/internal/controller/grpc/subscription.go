package grpc

import (
	"context"
	"errors"

	subscriberpb "repo-stat/proto/subscriber"
	"repo-stat/subscriber/internal/usecase"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateSubscription(
	ctx context.Context,
	req *subscriberpb.CreateSubscriptionRequest,
) (*subscriberpb.CreateSubscriptionResponse, error) {
	if req.GetSubscription() == nil {
		return nil, status.Error(codes.InvalidArgument, "subscription is required")
	}

	subscription, err := s.subscription.Create(ctx, req.GetSubscription().GetOwner(), req.GetSubscription().GetRepoName())
	if err != nil {
		return nil, toStatusError(err)
	}

	return &subscriberpb.CreateSubscriptionResponse{
		Subscription: &subscriberpb.Subscription{
			Owner:    subscription.Owner,
			RepoName: subscription.RepoName,
		},
	}, nil
}

func (s *Server) DeleteSubscription(
	ctx context.Context,
	req *subscriberpb.DeleteSubscriptionRequest,
) (*subscriberpb.DeleteSubscriptionResponse, error) {
	if req.GetSubscription() == nil {
		return nil, status.Error(codes.InvalidArgument, "subscription is required")
	}

	if err := s.subscription.Delete(ctx, req.GetSubscription().GetOwner(), req.GetSubscription().GetRepoName()); err != nil {
		return nil, toStatusError(err)
	}

	return &subscriberpb.DeleteSubscriptionResponse{}, nil
}

func (s *Server) ListSubscriptions(
	ctx context.Context,
	_ *subscriberpb.ListSubscriptionsRequest,
) (*subscriberpb.ListSubscriptionsResponse, error) {
	subscriptions, err := s.subscription.List(ctx)
	if err != nil {
		return nil, toStatusError(err)
	}

	result := make([]*subscriberpb.Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, &subscriberpb.Subscription{
			Owner:    subscription.Owner,
			RepoName: subscription.RepoName,
		})
	}

	return &subscriberpb.ListSubscriptionsResponse{Subscriptions: result}, nil
}

func toStatusError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrOwnerRequired), errors.Is(err, usecase.ErrRepoNameRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, pgx.ErrNoRows):
		return status.Error(codes.NotFound, "subscription not found")
	case errors.Is(err, usecase.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, "subscription already exists")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
