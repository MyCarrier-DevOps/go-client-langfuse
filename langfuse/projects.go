package langfuse

import (
	"context"
	"net/http"
)

// ProjectsService handles operations related to projects
type ProjectsService service

// Project represents a project in langfuse
type Project struct {
	ID            string
	Metadata      map[string]interface{}
	Name          string
	RetentionDays int
}

// GetProject retrieves a project associated with the given API token
// https://api.reference.langfuse.com/#tag/projects/get/api/public/projects
func (s *ProjectsService) GetProject(ctx context.Context) (*Project, *http.Response, error) {
	u := "/api/public/projects"
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var project Project
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, resp, err
	}
	return &project, resp, nil
}
