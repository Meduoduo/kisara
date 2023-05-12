package docker

import "github.com/Yeuoly/kisara/src/types"

func launchX86Qemu(
	docker *Docker,
	image_name string, uid int, protocol string, port int,
	subnet_names []string, limit types.KisaraVmLimit,
) {
	docker.CreateContainer()
}
