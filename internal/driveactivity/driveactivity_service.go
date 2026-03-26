package driveactivity

import (
	"context"

	"google.golang.org/api/driveactivity/v2"
)

// DriveActivityService defines the interface for Google Drive Activity API operations.
// This interface enables dependency injection and testing with mocks.
type DriveActivityService interface {
	// QueryActivity queries drive activity with the given request parameters.
	QueryActivity(ctx context.Context, req *driveactivity.QueryDriveActivityRequest) (*driveactivity.QueryDriveActivityResponse, error)
}

// RealDriveActivityService wraps the Drive Activity API client and implements DriveActivityService.
type RealDriveActivityService struct {
	service *driveactivity.Service
}

// NewRealDriveActivityService creates a new RealDriveActivityService wrapping the given API service.
func NewRealDriveActivityService(service *driveactivity.Service) *RealDriveActivityService {
	return &RealDriveActivityService{service: service}
}

// QueryActivity queries drive activity with the given request parameters.
func (s *RealDriveActivityService) QueryActivity(ctx context.Context, req *driveactivity.QueryDriveActivityRequest) (*driveactivity.QueryDriveActivityResponse, error) {
	return s.service.Activity.Query(req).Context(ctx).Do()
}
