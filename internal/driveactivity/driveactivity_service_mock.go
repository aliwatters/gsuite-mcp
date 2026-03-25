package driveactivity

import (
	"context"

	"google.golang.org/api/driveactivity/v2"
)

// MockDriveActivityService implements DriveActivityService for testing.
type MockDriveActivityService struct {
	// Activities stores mock activity data returned by QueryActivity.
	Activities []*driveactivity.DriveActivity

	// Errors allows tests to configure specific errors for methods.
	Errors struct {
		QueryActivity error
	}

	// Calls tracks method invocations for verification.
	Calls struct {
		QueryActivity []*driveactivity.QueryDriveActivityRequest
	}
}

// NewMockDriveActivityService creates a new mock Drive Activity service with default test data.
func NewMockDriveActivityService() *MockDriveActivityService {
	m := &MockDriveActivityService{}

	m.Activities = []*driveactivity.DriveActivity{
		{
			PrimaryActionDetail: &driveactivity.ActionDetail{
				Edit: &driveactivity.Edit{},
			},
			Actors: []*driveactivity.Actor{
				{
					User: &driveactivity.User{
						KnownUser: &driveactivity.KnownUser{
							PersonName:    "people/12345",
							IsCurrentUser: true,
						},
					},
				},
			},
			Targets: []*driveactivity.Target{
				{
					DriveItem: &driveactivity.DriveItem{
						Name:     "items/abc123",
						Title:    "Test Document",
						MimeType: "application/vnd.google-apps.document",
					},
				},
			},
			Timestamp: "2024-01-15T10:00:00.000Z",
		},
		{
			PrimaryActionDetail: &driveactivity.ActionDetail{
				Create: &driveactivity.Create{},
			},
			Actors: []*driveactivity.Actor{
				{
					User: &driveactivity.User{
						KnownUser: &driveactivity.KnownUser{
							PersonName: "people/67890",
						},
					},
				},
			},
			Targets: []*driveactivity.Target{
				{
					DriveItem: &driveactivity.DriveItem{
						Name:     "items/def456",
						Title:    "New Spreadsheet",
						MimeType: "application/vnd.google-apps.spreadsheet",
					},
				},
			},
			Timestamp: "2024-01-14T09:30:00.000Z",
		},
	}

	return m
}

// QueryActivity returns mock activity data.
func (m *MockDriveActivityService) QueryActivity(ctx context.Context, req *driveactivity.QueryDriveActivityRequest) (*driveactivity.QueryDriveActivityResponse, error) {
	m.Calls.QueryActivity = append(m.Calls.QueryActivity, req)

	if m.Errors.QueryActivity != nil {
		return nil, m.Errors.QueryActivity
	}

	// Filter by item name if specified
	activities := m.Activities
	if req.ItemName != "" {
		var filtered []*driveactivity.DriveActivity
		for _, a := range activities {
			for _, t := range a.Targets {
				if t.DriveItem != nil && t.DriveItem.Name == req.ItemName {
					filtered = append(filtered, a)
					break
				}
			}
		}
		activities = filtered
	}

	if len(activities) == 0 && req.ItemName != "" {
		// Return empty result (not an error) for unfound items
		return &driveactivity.QueryDriveActivityResponse{
			Activities: []*driveactivity.DriveActivity{},
		}, nil
	}

	return &driveactivity.QueryDriveActivityResponse{
		Activities: activities,
	}, nil
}
