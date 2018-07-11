package version

import "fmt"

var (
	// Version app version
	Version = "0.1.0" // set on build time
	// Revision git commit short commithash
	Revision = "xxxxxx" // set on build time
)

// FullVersion return sackerel full version text
func FullVersion() string {
	return fmt.Sprintf("%s, build %s", Version, Revision)
}
