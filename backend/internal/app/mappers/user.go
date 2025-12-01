package mappers

import (
	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserToProto(user domain.User) *desc.User {
	userProto := &desc.User{
		Id:                     user.ID.String(),
		TelegramId:             int32(user.TelegramID),
		Username:               user.TelegramUsername.V,
		DateOfBirth:            timestamppb.New(user.DateOfBirth),
		FirstName:              user.FirstName.V,
		LastName:               user.LastName.V,
		Weight:                 user.Weight.V,
		Height:                 user.Height.V,
		ProfilePictureUrl:      user.ProfilePicURL.V,
		CreatedAt:              timestamppb.New(user.CreatedAt),
		UpdatedAt:              timestamppb.New(user.UpdatedAt),
		HasCompletedOnboarding: user.HasCompletedOnboarding,
	}
	if !user.DateOfBirth.IsZero() {
		userProto.DateOfBirth = timestamppb.New(user.DateOfBirth)
	}

	return userProto
}

func UserFactToProto(fact domain.UserFact) *desc.UserFact {
	return &desc.UserFact{
		Id:        fact.ID.String(),
		UserId:    fact.UserID.String(),
		Content:   fact.Content,
		CreatedAt: timestamppb.New(fact.CreatedAt),
		UpdatedAt: timestamppb.New(fact.UpdatedAt),
	}
}

func UserFactsToProto(facts []domain.UserFact) []*desc.UserFact {
	result := make([]*desc.UserFact, 0, len(facts))
	for _, fact := range facts {
		result = append(result, UserFactToProto(fact))
	}
	return result
}
