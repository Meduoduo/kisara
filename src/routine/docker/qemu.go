package docker

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/types"
)

func launchX86Qemu(
	docker *Docker,
	image_id string, uid int, protocol_port string,
	subnet_names []string, limit types.KisaraVmLimit,
) (*types.Container, error) {
	// TODO: check if vm image exists and pull it

	// TODO: fetch vm image file path
	path := fmt.Sprintf("/usr/local/kisara/storage/vm/%s/", image_id)
	image := types.KisaraVMImage{}

	return docker.CreateContainer(
		"yeuoly/kisara-vm-qemu-x86:latest",
		uid, protocol_port,
		subnet_names, "vm-qemu-x86",
		map[string]string{
			"QEMU_ENV": "ENV",
		},
		map[string]string{
			path: "/var/qemu/vm/",
		},
		image.Limit.Cpu, image.Limit.Mem, image.Limit.Disk,
	)
}
