package docker

import (
	"errors"

	"github.com/Yeuoly/kisara/src/types"
)

/*
	as we know, docker base on linux kernel, so containers can use linux kernel's resource
	but if we want to launch a virtual machine, common docker is not enough
	but.. qemu can help us
	we can use docker container to host qemu, and use qemu to launch a virtual machine
	this will not be slower than common qemu, and container cost just a little bit of resource
	actualy, we can consider qemu run in host directly even if we use container

	without qemu, there is kvm we can use, but to use kvm, we need to run docker in privileged mode
	and it only supports 5.4+ kernel which is not common in most linux distribution
	so we use qemu instead of kvm
*/

type VirtualMachineInf interface {
	// init vm
	Init(docker *Docker) error
	// launch vm
	LaunchVm(docker *Docker, image_id string, networks []string, limit types.KisaraVmLimit) (*types.VM, error)
	// stop vm
	StopVm(docker *Docker, vm_id string) error
	// list vm
	ListVm(docker *Docker) ([]*types.VM, error)
	// get vm
	GetVm(docker *Docker, vm_id string) (*types.VM, error)
}

var (
	ErrVmNotFound = errors.New("VM not found")
)
