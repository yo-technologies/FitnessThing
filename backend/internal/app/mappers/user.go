package mappers

import (
	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserToProto(user domain.User) *desc.User {
	userProto := &desc.User{
		Id:                user.ID.String(),
		FirstName:         user.FirstName.V,
		LastName:          user.LastName.V,
		Weight:            user.Weight.V,
		Height:            user.Height.V,
		ProfilePictureUrl: user.ProfilePicURL.V,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
	}
	if !user.DateOfBirth.IsZero() {
		userProto.DateOfBirth = timestamppb.New(user.DateOfBirth)
	}

	return userProto
}
