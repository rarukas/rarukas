package server

const (
	// RarukasDefaultHTTPPort is number of default health check port
	RarukasDefaultHTTPPort = 8080
	// RarukasDefaultSSHPort is number of default ssh server port
	RarukasDefaultSSHPort = 2222
	// RarukasPublicKeyEnv is the key name of the environment variable used to pass ssh-public-keys
	RarukasPublicKeyEnv = "RARUKAS_PUBLIC_KEY"
	// RarukasCommandEnv is the key name of the environment variable used to pass container command
	RarukasCommandEnv = "RARUKAS_COMMAND"
)
