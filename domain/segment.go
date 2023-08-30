package domain

import (
	"context"
	"errors"
	"fmt"
)

type SegmentStorage interface {
	CreateSegment(ctx context.Context, name string) error
	DeleteSegment(ctx context.Context, name string) error
	AddUserToSegment(ctx context.Context, user int, segments []string) error
	DeleteUserFromSegment(ctx context.Context, user int, segments []string) error
	GetUserSegments(ctx context.Context, user int) ([]string, error)
}

var (
	//для этой ошибки выводить какой сегмент не был найден
	ErrSegmentNotFound             = errors.New("can't find the segment")
	ErrSegmentIsAlreadyExists      = errors.New("segment with this name is already exists")
	ErrUserIsAlreadyHasThisSegment = errors.New("user is already has this segment")
	//для этой ошибки выводить какого семента не было у пользователя
	ErrUserHaveNotThisSegment = errors.New("user doesn't have this segment")
)

type SegmentService struct {
	storage SegmentStorage
}

func NewSegmentService(storage SegmentStorage) (ss SegmentService) {
	return SegmentService{
		storage: storage,
	}
}

func (ss *SegmentService) CreateSegment(ctx context.Context, name string) error {
	err := ss.storage.CreateSegment(ctx, name)
	if err != nil {
		return fmt.Errorf("creating segment: %w", err)
	}
	return nil
}

func (ss *SegmentService) DeleteSegment(ctx context.Context, name string) error {
	err := ss.storage.DeleteSegment(ctx, name)
	if err != nil {
		return fmt.Errorf("deleting segment: %w", err)
	}
	return nil
}

func (ss *SegmentService) ChangeUserSegments(ctx context.Context, user int, segmentsToAdd []string, segmentsToDelete []string) error {
	var errs error
	if len(segmentsToAdd) != 0 {
		err := ss.storage.AddUserToSegment(ctx, user, segmentsToAdd)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("adding user to segments: %w", err))
		}
	}

	if len(segmentsToDelete) != 0 {
		err := ss.storage.DeleteUserFromSegment(ctx, user, segmentsToDelete)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("deleting user from segments: %w", err))
		}
	}

	return errs
}

func (ss *SegmentService) GetUserSegments(ctx context.Context, user int) (segmnets []string, err error) {
	segments, err := ss.storage.GetUserSegments(ctx, user)
	if err != nil {
		return []string{}, fmt.Errorf("getting segments: %w", err)
	}

	return segments, nil
}
