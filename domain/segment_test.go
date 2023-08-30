package domain

import (
	"context"
	"errors"
	"testing"
)

func TestSegmentService_ChangeUserSegments(t *testing.T) {
	type fields struct {
		storage *storageMock
	}
	type args struct {
		ctx              context.Context
		user             int
		segmentsToAdd    []string
		segmentsToDelete []string
	}
	tests := []struct {
		name       string
		beforeTest func(t *testing.T, f *fields)
		afterTest  func(t *testing.T, f *fields)
		args       args
		wantErr    bool
	}{
		{
			name: "given error from AddUserToSegment return an error",
			args: args{
				ctx:              context.Background(),
				user:             0,
				segmentsToAdd:    []string{"TEST_SEGMENT"},
				segmentsToDelete: []string{},
			},
			beforeTest: func(t *testing.T, f *fields) {
				f.storage.AddUserToSegmentFunc = func(ctx context.Context, user int, segments []string) error {
					return errors.New("fail")
				}
			},
			afterTest: func(t *testing.T, f *fields) {
				if len(f.storage.AddUserToSegmentCalls) != 1 {
					t.Errorf("Expected 1 call to storage.AddUserToSegment, but got %d", len(f.storage.AddUserToSegmentCalls))
				}
			},
			wantErr: true,
		},

		{
			name: "given error from DeleteUserFromSegment return an error",
			args: args{
				ctx:              context.Background(),
				user:             0,
				segmentsToAdd:    []string{},
				segmentsToDelete: []string{"TEST_SEGMENT"},
			},
			beforeTest: func(t *testing.T, f *fields) {
				f.storage.DeleteUserFromSegmentFunc = func(ctx context.Context, user int, segments []string) error {
					return errors.New("fail")
				}
			},
			afterTest: func(t *testing.T, f *fields) {
				if len(f.storage.DeleteUserFromSegmentCalls) != 1 {
					t.Errorf("Expected 1 call to storage.DeleteUserFromSegment, but got %d", len(f.storage.DeleteUserFromSegmentCalls))
				}
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				ctx:              context.Background(),
				user:             0,
				segmentsToAdd:    []string{"TEST_SEGMENT"},
				segmentsToDelete: []string{"TEST_SEGMENT"},
			},
			beforeTest: func(t *testing.T, f *fields) {
				f.storage.AddUserToSegmentFunc = func(ctx context.Context, user int, segments []string) error {
					return nil
				}
				f.storage.DeleteUserFromSegmentFunc = func(ctx context.Context, user int, segments []string) error {
					return nil
				}
			},
			afterTest: func(t *testing.T, f *fields) {
				if len(f.storage.AddUserToSegmentCalls) != 1 {
					t.Errorf("Expected 1 call to storage.AddUserToSegment, but got %d", len(f.storage.AddUserToSegmentCalls))
				}
				if len(f.storage.DeleteUserFromSegmentCalls) != 1 {
					t.Errorf("Expected 1 call to storage.DeleteUserFromSegment, but got %d", len(f.storage.DeleteUserFromSegmentCalls))
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fields{
				storage: &storageMock{
					AddUserToSegmentFunc: func(ctx context.Context, user int, segments []string) error {
						return nil
					},
					DeleteUserFromSegmentFunc: func(ctx context.Context, user int, segments []string) error {
						return nil
					},
				},
			}

			if tt.beforeTest != nil {
				tt.beforeTest(t, &f)
			}

			ss := NewSegmentService(f.storage)
			if err := ss.ChangeUserSegments(tt.args.ctx, tt.args.user, tt.args.segmentsToAdd, tt.args.segmentsToDelete); (err != nil) != tt.wantErr {
				t.Errorf("SegmentService.ChangeUserSegments() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.afterTest != nil {
				tt.afterTest(t, &f)
			}
		})
	}
}
