package arukas

import (
	"encoding/json"
	"strconv"
	"strings"
)

// LinkID used when creating new app/service
const LinkID = 1

const (
	// TypeApps represents the "apps" type
	TypeApps = "apps"
	// TypeServices represents the "services" type
	TypeServices = "services"
	// TypeServicePlans represents the "service_plans" type
	TypeServicePlans = "service-plans"
	// TypeUsers represents the "users" type
	TypeUsers = "users"
	// TypeRegions represents the "regions" type
	TypeRegions = "regions"
)

const (
	// StatusBooting represents the "booting" status
	StatusBooting = "booting"
	// StatusTerminated represents the "terminated" status
	StatusTerminated = "terminated"
	// StatusRunning represents the "running" status
	StatusRunning = "running"
	// StatusStopping represents the "stopping" status
	StatusStopping = "stopping"
	// StatusStopped represents the "stopped" status
	StatusStopped = "stopped"
	// StatusRebooting represents the "rebooting" status
	StatusRebooting = "rebooting"
)

// Port represents the port information you want expose in container
type Port struct {
	// Protocol is either tcp/udp
	Protocol string `json:"protocol"`
	// Number is port number that container exposes.
	// Valid value range is [1 - 65535]
	Number int32 `json:"number"`
}

// Ports is a slice of Ports. A service can have multiple ports.
type Ports []*Port
type oldPortFormat Ports
type newPortFormat []string

func (pf newPortFormat) toPorts() (Ports, error) {
	ports := make(Ports, 0)
	for _, p := range pf {
		var (
			parsedPort *Port
			err        error
		)
		if parsedPort, err = parseNewPortFormat(p); err != nil {
			return nil, err
		}

		ports = append(ports, parsedPort)
	}

	return ports, nil
}

func (pf oldPortFormat) toPorts() (Ports, error) {
	ports := make(Ports, 0)
	for _, p := range pf {
		ports = append(ports,
			&Port{
				Protocol: p.Protocol,
				Number:   p.Number,
			},
		)
	}
	return ports, nil
}

func parseNewPortFormat(str string) (*Port, error) {
	var protocol string
	var parsedInt int64
	var number int32
	var err error
	splitted := strings.Split(str, "/")
	if len(splitted) <= 1 {
		protocol = "tcp"
	} else {
		protocol = splitted[1]
	}

	if parsedInt, err = strconv.ParseInt(splitted[0], 10, 32); err != nil {
		return nil, err
	}
	number = int32(parsedInt)

	return &Port{Protocol: protocol, Number: number}, nil
}

// UnmarshalJSON parses ports in both old and new port format and convert them to Port.
func (ports *Ports) UnmarshalJSON(data []byte) error {
	op := new(oldPortFormat)
	np := new(newPortFormat)
	var err error
	if err = json.Unmarshal(data, op); err != nil {
		// Attempt parse as new format
		if err = json.Unmarshal(data, np); err != nil {
			return err
		}
		if *ports, err = np.toPorts(); err != nil {
			return err
		}
	} else {
		if *ports, err = op.toPorts(); err != nil {
			return err
		}
	}
	return nil
}

// ValidProtocols is a list of valid protocol
var ValidProtocols = []string{"tcp", "udp"}

// PortMapping represents actual port mapping information on running container
type PortMapping struct {
	Host          string `json:"host"`
	Protocol      string `json:"protocol"`
	ContainerPort int32  `json:"container-port"`
	ServicePort   int32  `json:"service-port"`
}

// CustomDomain represents custom_domain object
type CustomDomain struct {
	Name string `json:"name"`
}

// CustomDomains returns array of *CustomDomain created from domain name list
func CustomDomains(domains ...string) []*CustomDomain {
	res := []*CustomDomain{}
	for _, d := range domains {
		res = append(res, &CustomDomain{Name: d})
	}
	return res
}

// Env represents environment object
type Env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
