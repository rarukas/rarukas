package arukas

import (
	"encoding/json"
	"time"
)

// AppListData represents data object(included []app and child services)
type AppListData struct {
	Data     []*App        `json:"data"`
	Included []interface{} `json:"included,omitempty"`
}

// App represents a app object
type App struct {
	ID            string           `json:"id"`
	Type          string           `json:"type,omitempty"`
	Attributes    *AppAttr         `json:"attributes,omitempty"`
	Relationships *AppRelationship `json:"relationships,omitempty"`
}

// AppID returns data.id
func (a *App) AppID() string {
	return a.ID
}

// Name returns data.attributes.Name
func (a *App) Name() string {
	return a.Attributes.Name
}

// CreatedAt returns data.attributes.created_at
func (a *App) CreatedAt() *time.Time {
	return a.Attributes.CreatedAt
}

// UpdatedAt returns data.attributes.updated_at
func (a *App) UpdatedAt() *time.Time {
	return a.Attributes.UpdatedAt
}

// ServiceID returns service id
func (a *App) ServiceID() string {
	return a.Relationships.Services.Data[0].ID
}

// AppAttr represents app.attributes object
type AppAttr struct {
	Name      string     `json:"name,omitempty"`
	CreatedAt *time.Time `json:"created-at,omitempty"`
	UpdatedAt *time.Time `json:"updated-at,omitempty"`
}

// AppData represents data object(included app and child service)
type AppData struct {
	Data     *App          `json:"data"`
	Included []interface{} `json:"included,omitempty"`
}

// AppID returns data.id
func (a *AppData) AppID() string {
	return a.Data.ID
}

// Type returns data.type
func (a *AppData) Type() string {
	return a.Data.Type
}

// Name returns data.attributes.Name
func (a *AppData) Name() string {
	return a.Data.Name()
}

// CreatedAt returns data.attributes.created_at
func (a *AppData) CreatedAt() *time.Time {
	return a.Data.CreatedAt()
}

// UpdatedAt returns data.attributes.updated_at
func (a *AppData) UpdatedAt() *time.Time {
	return a.Data.UpdatedAt()
}

// ServiceID returns service id
func (a *AppData) ServiceID() string {
	return a.Data.ServiceID()
}

// Service returns service(from included)
func (a *AppData) Service() *Service {
	for _, v := range a.Included {
		var service Service
		data, err := json.Marshal(v)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &service); err != nil {
			continue
		}

		return &service
	}
	return nil
}
