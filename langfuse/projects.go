package langfuse

type Project struct {
	ID            string
	Metadata      map[string]interface{}
	Name          string
	RetentionDays int
}
