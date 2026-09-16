package domain

type SCIMUser struct {
	Schemas    []string    `json:"schemas"`
	ID         string      `json:"id"`
	ExternalID string      `json:"externalId,omitempty"`
	UserName   string      `json:"userName"`
	Active     bool        `json:"active"`
	Emails     []SCIMEmail `json:"emails"`
	Meta       SCIMMeta    `json:"meta"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
	Type    string `json:"type,omitempty"`
}

type SCIMMeta struct {
	ResourceType string `json:"resourceType"`
	Created      string `json:"created"`
	LastModified string `json:"lastModified"`
	Location     string `json:"location"`
}

type SCIMListResponse struct {
	Schemas      []string    `json:"schemas"`
	TotalResults int64       `json:"totalResults"`
	ItemsPerPage int         `json:"itemsPerPage"`
	StartIndex   int         `json:"startIndex"`
	Resources    []*SCIMUser `json:"Resources"`
}

type SCIMServiceProviderConfig struct {
	Schemas []string `json:"schemas"`
	Patch   struct {
		Supported bool `json:"supported"`
	} `json:"patch"`
	Bulk struct {
		Supported bool `json:"supported"`
	} `json:"bulk"`
	Filter struct {
		Supported bool `json:"supported"`
	} `json:"filter"`
	ChangePassword struct {
		Supported bool `json:"supported"`
	} `json:"changePassword"`
	Sort struct {
		Supported bool `json:"supported"`
	} `json:"sort"`
	Etag struct {
		Supported bool `json:"supported"`
	} `json:"etag"`
	AuthenticationSchemes []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	} `json:"authenticationSchemes"`
}