package user

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) ListUserFacts(ctx context.Context, in *desc.ListUserFactsRequest) (*desc.ListUserFactsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.ListUserFacts")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	facts, err := i.service.ListUserFacts(ctx, userID)
	if err != nil {
		logger.Errorf("error listing user facts: %v", err)
		return nil, err
	}

	return &desc.ListUserFactsResponse{Facts: mappers.UserFactsToProto(facts)}, nil
}

func (i *Implementation) DeleteUserFact(ctx context.Context, in *desc.DeleteUserFactRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.DeleteUserFact")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	factID, err := domain.ParseID(in.GetFactId())
	if err != nil {
		logger.Errorf("error parsing fact id: %v", err)
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	if err := i.service.DeleteUserFact(ctx, userID, factID); err != nil {
		logger.Errorf("error deleting user fact: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
