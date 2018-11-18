package ec2instances

// InstanceSize represent cpu and memory of a instance
type InstanceSize struct {
	InstanceType string  `json:"instance_type"`
	VCPU         int     `json:"vCPU"`
	Memory       float64 `json:"memory"`
}

//go:generate go run gen.go
