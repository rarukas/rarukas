package arukas

// RelationshipData represents relationship data
type RelationshipData struct {
	Data *Relationship `json:"data,omitempty"`
}

// Relationship represents relationship body
type Relationship struct {
	ID     string `json:"id,omitempty"`
	TempID int32  `json:"temp_id,omitempty"`
	Type   string `json:"type,omitempty"`
}

// AppRelationship represents App relationship data
type AppRelationship struct {
	User    *RelationshipData `json:"user,omitempty"`
	Service *RelationshipData `json:"service,omitempty"`
}

// ServiceRelationship represents Service relationship data
type ServiceRelationship struct {
	App         *RelationshipData `json:"app,omitempty"`
	ServicePlan *RelationshipData `json:"service_plan,omitempty"`
}

// NewAppRelationship creates new AppRelationship with default values
func NewAppRelationship() *AppRelationship {
	return &AppRelationship{
		Service: &RelationshipData{
			Data: &Relationship{
				TempID: TempID,
				Type:   TypeServices,
			},
		},
	}

}

// NewServiceRelationship creates new ServiceRelationship with PlanID
func NewServiceRelationship(region, plan string) *ServiceRelationship {
	if region == "" || plan == "" {
		return nil
	}
	return &ServiceRelationship{
		ServicePlan: &RelationshipData{
			Data: &Relationship{
				ID:   PlanID(region, plan),
				Type: TypeServicePlans,
			},
		},
	}
}
