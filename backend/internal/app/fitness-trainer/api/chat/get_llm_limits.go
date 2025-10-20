package chat

import (
	"context"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) GetLLMLimits(ctx context.Context, _ *emptypb.Empty) (*desc.GetLLMLimitsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.GetLLMLimits")
	defer span.Finish()

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, domain.ErrInternal
	}

	limits, err := i.service.GetLLMLimits(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &desc.GetLLMLimitsResponse{
		DailyLimit: int32(limits.DailyLimit),
		Used:       int32(limits.Used),
		Reserved:   int32(limits.Reserved),
		Remaining:  int32(limits.Remaining),
	}, nil
}
