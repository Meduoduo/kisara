package docker

import (
	"sync"

	"github.com/Yeuoly/kisara/src/types"
)

type VirtualMachine struct {
	VirtualMachineInf

	// vm map
	vm_map      map[string]*types.VM
	vm_map_lock sync.Mutex
}

var (
	sington_vm *VirtualMachine
)

func GetVirtualMachine() *VirtualMachine {
	if sington_vm == nil {
		sington_vm = new(VirtualMachine)
		sington_vm.Init(nil)
	}
	return sington_vm
}

func (vm *VirtualMachine) Init(docker *Docker) error {
	vm.vm_map = make(map[string]*types.VM)
	return nil
}

func (vm *VirtualMachine) LaunchVm(docker *Docker, image_id string, networks []string, limit types.KisaraVmLimit) (*types.VM, error) {
	return nil, nil
}

func (vm *VirtualMachine) StopVm(docker *Docker, vm_id string) error {
	vm.vm_map_lock.Lock()
	instance := vm.vm_map[vm_id]
	vm.vm_map_lock.Unlock()

	if instance == nil {
		return ErrVmNotFound
	}

	// stop vm

	return nil
}

func (vm *VirtualMachine) ListVm(docker *Docker) ([]*types.VM, error) {
	vm.vm_map_lock.Lock()
	defer vm.vm_map_lock.Unlock()

	list := make([]*types.VM, 0)
	for _, v := range vm.vm_map {
		list = append(list, v)
	}
	return list, nil
}
