package arukas

// RelationshipData represents relationship data
type RelationshipData struct {
	Data *Relationship `json:"data,omitempty"`
}

// RelationshipDataList represents relationship data(list)
type RelationshipDataList struct {
	Data []*Relationship `json:"data,omitempty"`
}

// Relationship represents relationship body
type Relationship struct {
	ID     string `json:"id,omitempty"`
	LinkID int32  `json:"lid,omitempty"`
	Type   string `json:"type,omitempty"`
}

// AppRelationship represents App relationship data
type AppRelationship struct {
	User     *RelationshipData     `json:"user,omitempty"`
	Services *RelationshipDataList `json:"services,omitempty"`
}

// ServiceRelationship represents Services relationship data
type ServiceRelationship struct {
	App         *RelationshipData `json:"app,omitempty"`
	ServicePlan *RelationshipData `json:"service-plan,omitempty"`
}

// NewAppRelationship creates new AppRelationship with default values
func NewAppRelationship() *AppRelationship {
	return &AppRelationship{
		Services: &RelationshipDataList{
			Data: []*Relationship{
				{
					LinkID: LinkID,
					Type:   TypeServices,
				},
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
