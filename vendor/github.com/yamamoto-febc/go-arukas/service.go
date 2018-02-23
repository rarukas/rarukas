package arukas

import "time"

// ServiceListData represents services data
type ServiceListData struct {
	Data []*Service `json:"data"`
}

// Service represents service object
type Service struct {
	ID            string               `json:"id"`
	TempID        int32                `json:"temp_id,omitempty"`
	Type          string               `json:"type,omitempty"`
	Attributes    *ServiceAttr         `json:"attributes,omitempty"`
	Relationships *ServiceRelationship `json:"relationships,omitempty"`
}

// AppID returns data.attributes.app_id
func (s *Service) AppID() string {
	return s.Attributes.AppID
}

// Image returns data.attributes.image
func (s *Service) Image() string {
	return s.Attributes.Image
}

// Command returns data.attributes.command
func (s *Service) Command() string {
	return s.Attributes.Command
}

// Instances returns data.attributes.instances
func (s *Service) Instances() int32 {
	return s.Attributes.Instances
}

// CPUs returns data.attributes.cups
func (s *Service) CPUs() float32 {
	return s.Attributes.CPUs
}

// Memory returns data.attributes.memory
func (s *Service) Memory() int32 {
	return s.Attributes.Memory
}

// Environment returns data.attributes.environment
func (s *Service) Environment() []*Env {
	return s.Attributes.Environment
}

// Ports returns data.attributes.ports
func (s *Service) Ports() []*Port {
	return s.Attributes.Ports
}

// PortMappings returns data.attributes.port_mappings
func (s *Service) PortMappings() [][]*PortMapping {
	return s.Attributes.PortMappings
}

// PortMapping returns data.attributes.port_mappings[0]
func (s *Service) PortMapping() []*PortMapping {
	if len(s.Attributes.PortMappings) == 0 {
		return nil
	}
	return s.Attributes.PortMappings[0]
}

// CreatedAt returns data.attributes.created_at
func (s *Service) CreatedAt() *time.Time {
	return s.Attributes.CreatedAt
}

// UpdatedAt returns data.attributes.updated_at
func (s *Service) UpdatedAt() *time.Time {
	return s.Attributes.UpdatedAt
}

// Status returns data.attributes.status
func (s *Service) Status() string {
	return s.Attributes.Status
}

// SubDomain returns data.attributes.subdomain
func (s *Service) SubDomain() string {
	return s.Attributes.SubDomain
}

// EndPoint returns data.attributes.endpoint
func (s *Service) EndPoint() string {
	return s.Attributes.EndPoint
}

// PlanID returns data.relationship.service_plan.data.id
func (s *Service) PlanID() string {
	return s.Relationships.ServicePlan.Data.ID
}

// ServiceAttr represents service.attributes object
type ServiceAttr struct {
	AppID                    string           `json:"app_id,omitempty"`
	Image                    string           `json:"image"`
	Command                  string           `json:"command"`
	Instances                int32            `json:"instances"`
	CPUs                     float32          `json:"cups,omitempty"`
	Memory                   int32            `json:"memory,omitempty"`
	Environment              []*Env           `json:"environment"`
	Ports                    []*Port          `json:"ports,omitempty"`
	PortMappings             [][]*PortMapping `json:"port_mappings,omitempty"`
	CreatedAt                *time.Time       `json:"created_at,omitempty"`
	UpdatedAt                *time.Time       `json:"updated_at,omitempty"`
	Status                   string           `json:"status,omitempty"`
	SubDomain                string           `json:"subdomain,omitempty"`
	EndPoint                 string           `json:"endpoint,omitempty"`
	CustomDomains            []*CustomDomain  `json:"custom_domains,omitempty"`
	LastInstanceFailedAt     *time.Time       `json:"last_instance_failed_at,omitempty"`
	LastInstanceFailedStatus string           `json:"last_instance_failed_status,omitempty"`
}

// ServiceData represents service data
type ServiceData struct {
	Data *Service `json:"data"`
}

// ServiceID returns data.id
func (s *ServiceData) ServiceID() string {
	return s.Data.ID
}

// AppID returns data.attributes.app_id
func (s *ServiceData) AppID() string {
	return s.Data.Attributes.AppID
}

// Type returns data.type
func (s *ServiceData) Type() string {
	return s.Data.Type
}

// Image returns data.attributes.image
func (s *ServiceData) Image() string {
	return s.Data.Image()
}

// Command returns data.attributes.command
func (s *ServiceData) Command() string {
	return s.Data.Command()
}

// Instances returns data.attributes.instances
func (s *ServiceData) Instances() int32 {
	return s.Data.Instances()
}

// CPUs returns data.attributes.cups
func (s *ServiceData) CPUs() float32 {
	return s.Data.CPUs()
}

// Memory returns data.attributes.memory
func (s *ServiceData) Memory() int32 {
	return s.Data.Memory()
}

// Environment returns data.attributes.environment
func (s *ServiceData) Environment() []*Env {
	return s.Data.Environment()
}

// Ports returns data.attributes.ports
func (s *ServiceData) Ports() []*Port {
	return s.Data.Ports()
}

// PortMappings returns data.attributes.port_mappings
func (s *ServiceData) PortMappings() [][]*PortMapping {
	return s.Data.PortMappings()
}

// PortMapping returns data.attributes.port_mappings[0]
func (s *ServiceData) PortMapping() []*PortMapping {
	return s.Data.PortMapping()
}

// CreatedAt returns data.attributes.created_at
func (s *ServiceData) CreatedAt() *time.Time {
	return s.Data.CreatedAt()
}

// UpdatedAt returns data.attributes.updated_at
func (s *ServiceData) UpdatedAt() *time.Time {
	return s.Data.UpdatedAt()
}

// Status returns data.attributes.status
func (s *ServiceData) Status() string {
	return s.Data.Status()
}

// SubDomain returns data.attributes.subdomain
func (s *ServiceData) SubDomain() string {
	return s.Data.SubDomain()
}

// EndPoint returns data.attributes.endpoint
func (s *ServiceData) EndPoint() string {
	return s.Data.EndPoint()
}

// PlanID returns data.relationship.service_plan.data.id
func (s *ServiceData) PlanID() string {
	return s.Data.PlanID()
}
