package types

type KisaraVmLimit struct {
	// CPU limit
	Cpu float64 `json:"cpu"` // 0.5 means 50% of one core
	// Memory limit
	Mem int64 `json:"mem"` // bytes
	// Disk limit
	Disk int64 `json:"disk"` // bytes
	// Network limit
	Network int64 `json:"network"` // bytes
}

type KisaraVmNetwork struct {
	// Network name
	Name string `json:"name"`
	// Network type
	Type string `json:"type"`
	// Ip
	Ip string `json:"ip"`
	// Gateway
	Gateway string `json:"gateway"`
	// Subnet
	Subnet string `json:"subnet"`
	// Mac
	Mac string `json:"mac"`
}

const (
	KISARA_VM_LIMIT_BYTES = 1
	KISARA_VM_LIMIT_KB    = 1024 * KISARA_VM_LIMIT_BYTES
	KISARA_VM_LIMIT_MB    = 1024 * KISARA_VM_LIMIT_KB
	KISARA_VM_LIMIT_GB    = 1024 * KISARA_VM_LIMIT_MB
	KISARA_VM_LIMIT_TB    = 1024 * KISARA_VM_LIMIT_GB
)

type KisaraVMImage struct {
	// Id of the VM image
	Id string `json:"id"`
	// Image name
	Image string `json:"image"`
	// Image tag
	Tag string `json:"tag"`
	// architecture
	Arch string `json:"arch"`
	// Type of VM
	Type string `json:"type"`
	// Base qemu image
	BaseImage string `json:"base_image"`
	// Image size
	Size int64 `json:"size"`
	// created time
	Created int64 `json:"created"`
	// checksum
	Checksum string `json:"checksum"` // sha256
}

const (
	KISARA_VM_ARCH_X86_64   = "x86_64"
	KISARA_VM_ARCH_x86      = "x86"
	KISARA_VM_ARCH_ARM      = "arm"
	KISARA_VM_ARCH_ARM64    = "arm64"
	KISARA_VM_ARCH_MIPS     = "mips"
	KISARA_VM_ARCH_MIPS64   = "mips64"
	KISARA_VM_ARCH_MIPS64LE = "mips64le"
	KISARA_VM_ARCH_MIPSLE   = "mipsle"

	// based on docker
	KISARA_VM_TYPE_QEMU = "qemu"
	KISARA_VM_TYPE_KVM  = "kvm"
	KISARA_VM_TYPE_LXC  = "lxc"
	KISARA_VM_TYPE_LXD  = "lxd"
)

type VM struct {
	// VM name
	Name string `json:"name"`
	// base image
	ImageId string `json:"image_id"`
	// VM type
	Type string `json:"type"`
	// VM arch
	Arch string `json:"arch"`
	// VM status
	Status string `json:"status"`
	// VM created time
	Created int64 `json:"created"`
	// Base  container id
	BaseContainerId string `json:"base_container_id"`
	// VM id
	Id string `json:"id"`
	// Limit of VM
	Limit KisaraVmLimit `json:"limit"`
	// Network
	Network []KisaraVmNetwork `json:"network"`
}

const (
	KISARA_VM_STATUS_RUNNING = "running"
	KISARA_VM_STATUS_STOPPED = "stopped"
)
