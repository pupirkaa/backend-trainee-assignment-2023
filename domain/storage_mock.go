package domain

import "context"

type storageMock struct {
	CreateSegmentFunc  func(ctx context.Context, name string) error
	CreateSegmentCalls []struct {
		ctx  context.Context
		name string
	}

	DeleteSegmentFunc  func(ctx context.Context, name string) error
	DeleteSegmentCalls []struct {
		ctx  context.Context
		name string
	}

	AddUserToSegmentFunc  func(ctx context.Context, user int, segments []string) error
	AddUserToSegmentCalls []struct {
		ctx      context.Context
		user     int
		segments []string
	}

	DeleteUserFromSegmentFunc  func(ctx context.Context, user int, segments []string) error
	DeleteUserFromSegmentCalls []struct {
		ctx      context.Context
		user     int
		segments []string
	}

	GetUserSegmentsFunc  func(ctx context.Context, user int) ([]string, error)
	GetUserSegmentsCalls []struct {
		ctx  context.Context
		user int
	}
}

func (m *storageMock) CreateSegment(ctx context.Context, name string) error {
	m.CreateSegmentCalls = append(m.CreateSegmentCalls, struct {
		ctx  context.Context
		name string
	}{
		ctx:  ctx,
		name: name,
	})
	return m.CreateSegmentFunc(ctx, name)
}
func (m *storageMock) DeleteSegment(ctx context.Context, name string) error {
	m.DeleteSegmentCalls = append(m.DeleteSegmentCalls, struct {
		ctx  context.Context
		name string
	}{
		ctx:  ctx,
		name: name,
	})
	return m.DeleteSegmentFunc(ctx, name)
}
func (m *storageMock) AddUserToSegment(ctx context.Context, user int, segments []string) error {
	m.AddUserToSegmentCalls = append(m.AddUserToSegmentCalls, struct {
		ctx      context.Context
		user     int
		segments []string
	}{
		ctx:      ctx,
		user:     user,
		segments: segments,
	})
	return m.AddUserToSegmentFunc(ctx, user, segments)
}
func (m *storageMock) DeleteUserFromSegment(ctx context.Context, user int, segments []string) error {
	m.DeleteUserFromSegmentCalls = append(m.DeleteUserFromSegmentCalls, struct {
		ctx      context.Context
		user     int
		segments []string
	}{
		ctx:      ctx,
		user:     user,
		segments: segments,
	})
	return m.DeleteUserFromSegmentFunc(ctx, user, segments)
}
func (m *storageMock) GetUserSegments(ctx context.Context, user int) ([]string, error) {
	m.GetUserSegmentsCalls = append(m.GetUserSegmentsCalls, struct {
		ctx  context.Context
		user int
	}{
		ctx:  ctx,
		user: user,
	})
	return m.GetUserSegmentsFunc(ctx, user)
}
