package arukas

// TempID used when creating new app/service
const TempID = 1

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
