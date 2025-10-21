package tags

type DefaultTags struct {
	CreationDate         string `json:"CreationDate"`
	Environment          string `json:"Environment"`
	InfraManagedBy       string `json:"InfraManagedBy"`
	ApplicationManagedBy string `json:"ApplicationManagedBy"`
	ProjectOwner         string `json:"ProjectOwner"`
	Project              string `json:"Project"`
	CreatedBy            string `json:"CreatedBy"`
	ProjectPrefix        string `json:"ProjectPrefix"`
	Suffix               string `json:"Suffix"`
}
