package langfuse

import (
	"encoding/json"
	"fmt"
)

// ProjectsService handles operations related to projects
type ProjectsService service

// GetProject retrieves a project associated with the given API token
// https://api.reference.langfuse.com/#tag/projects/get/api/public/projects
func (s *ProjectsService) GetProject() (map[string]interface{}, error) {
	u := "/api/public/projects"

	body, err := s.client.Do(u)
	if err != nil {
		return nil, fmt.Errorf("error fetching project: %w", err)
	}

	var appData map[string]interface{}
	err = json.Unmarshal(body, &appData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ArgoCD application data: %w", err)
	}

	return appData, nil
}
