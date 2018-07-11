package arukas

import (
	"github.com/hashicorp/go-multierror"
)

// RequestParam represents request parameter of Arukas.API
type RequestParam struct {
	// Name needs only when create app
	Name          string
	Command       string
	CustomDomains []string
	Image         string
	Instances     int32
	Ports         []*Port
	Environment   []*Env
	SubDomain     string
	Region        string
	Plan          string
}

func (p *RequestParam) validateRequired(requiredFields map[string]interface{}) error {
	var results error

	for k, v := range requiredFields {
		err := validateRequired(k, v)
		if err != nil {
			results = multierror.Append(results, err)
		}
	}
	return results
}

func (p *RequestParam) validateOptionals() error {
	var results error

	type rangeCheckField struct {
		value int
		min   int
		max   int
	}

	rangeCheckFields := map[string]rangeCheckField{
		"Instances": {
			value: int(p.Instances),
			min:   1,
			max:   10,
		},
		"Ports": {
			value: len(p.Ports),
			min:   1,
			max:   20,
		},
		"Environment": {
			value: len(p.Environment),
			min:   1,
			max:   20,
		},
	}

	for k, v := range rangeCheckFields {
		if v.value > 0 {
			err := validateRange(k, v.value, v.min, v.max)
			if err != nil {
				results = multierror.Append(results, err)
			}
		}
	}

	for _, port := range p.Ports {
		if err := validateInStrValues("Protocol", port.Protocol, ValidProtocols...); err != nil {
			results = multierror.Append(results, err)
		}
		if err := validateRange("Number", int(port.Number), 1, 65535); err != nil {
			results = multierror.Append(results, err)
		}
	}

	if err := valiateStrByteLen("Command", p.Command, 0, 4096); err != nil {
		results = multierror.Append(results, err)
	}

	if err := valiateStrByteLen("SubDomain", p.SubDomain, 0, 63); err != nil {
		results = multierror.Append(results, err)
	}

	if p.Region != "" {
		err := validateInStrValues("Region", p.Region, ValidRegions...)
		if err != nil {
			results = multierror.Append(results, err)
		}
	}

	if p.Plan != "" {
		err := validateInStrValues("Plan", p.Plan, ValidPlans...)
		if err != nil {
			results = multierror.Append(results, err)
		}
	}

	return results
}

// ValidateForCreate returns error if parameter is invalid
func (p *RequestParam) ValidateForCreate() error {

	var results error

	requiredFields := map[string]interface{}{
		"Name":      p.Name,
		"Image":     p.Image,
		"Plan":      p.Plan,
		"Instances": p.Instances,
		"Ports":     len(p.Ports),
	}

	if err := p.validateRequired(requiredFields); err != nil {
		results = multierror.Append(results, err)
	}

	if err := p.validateOptionals(); err != nil {
		results = multierror.Append(results, err)
	}

	return results
}

// ValidateForUpdate returns error if parameter is invalid
func (p *RequestParam) ValidateForUpdate() error {

	var results error

	requiredFields := map[string]interface{}{
		"Image":     p.Image,
		"Instances": p.Instances,
	}

	if err := p.validateRequired(requiredFields); err != nil {
		results = multierror.Append(results, err)
	}

	if err := p.validateOptionals(); err != nil {
		results = multierror.Append(results, err)
	}

	return results
}

// ToAppData returns *AppData built from RequestParam
func (p *RequestParam) ToAppData() *AppData {
	if err := p.ValidateForCreate(); err != nil {
		return nil
	}

	if p.Region == "" {
		p.Region = RegionJPTokyo
	}

	return &AppData{
		Data: &App{
			Type: TypeApps,
			Attributes: &AppAttr{
				Name: p.Name,
			},
			Relationships: NewAppRelationship(),
		},
		Included: []interface{}{
			&Service{
				Type:   TypeServices,
				LinkID: LinkID,
				Attributes: &ServiceAttr{
					Command:       p.Command,
					CustomDomains: CustomDomains(p.CustomDomains...),
					Image:         p.Image,
					Instances:     p.Instances,
					Ports:         p.Ports,
					Environment:   p.Environment,
					SubDomain:     p.SubDomain,
				},
				Relationships: NewServiceRelationship(p.Region, p.Plan),
			},
		},
	}

}

// ToServiceData returns *ServiceData built from RequestParam
func (p *RequestParam) ToServiceData() *ServiceData {
	// TODO Unknown where to set p.Name

	if err := p.ValidateForUpdate(); err != nil {
		return nil
	}
	if p.Region == "" {
		p.Region = RegionJPTokyo
	}

	return &ServiceData{
		Data: &Service{
			Attributes: &ServiceAttr{
				Image:         p.Image,
				Instances:     p.Instances,
				Command:       p.Command,
				CustomDomains: CustomDomains(p.CustomDomains...),
				Ports:         p.Ports,
				Environment:   p.Environment,
				SubDomain:     p.SubDomain,
			},
			Relationships: NewServiceRelationship(p.Region, p.Plan),
		},
	}

}
