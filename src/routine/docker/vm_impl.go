package docker

import (
	"sync"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	"github.com/Yeuoly/kisara/src/types"
	uuid "github.com/satori/go.uuid"
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

func (vm *VirtualMachine) LaunchVm(
	docker *Docker, image_id string, port_protocols string,
	networks []string,
	limit types.KisaraVmLimit,
) (*types.VM, error) {
	// use image_id to fetch image
	image := types.KisaraVMImage{}

	var container *types.Container
	var err error
	// launch qemu based on arch
	switch image.Arch {
	case types.KISARA_VM_ARCH_X86:
		container, err = launchX86Qemu(docker, image_id, 0, port_protocols, networks, limit)
	case types.KISARA_VM_ARCH_X86_64:
	case types.KISARA_VM_ARCH_ARM:
	case types.KISARA_VM_ARCH_ARM64:
	case types.KISARA_VM_ARCH_MIPS:
	case types.KISARA_VM_ARCH_MIPS64:
	case types.KISARA_VM_ARCH_MIPSLE:
	case types.KISARA_VM_ARCH_MIPS64LE:
	default:
		return nil, ErrVmNotSupportedArch
	}

	if err != nil {
		return nil, err
	}

	vm_instance := &types.VM{
		Name:            uuid.NewV4().String(),
		ImageId:         image_id,
		Type:            image.Type,
		Arch:            image.Arch,
		Status:          types.SERVICE_STATUS_RUNNING,
		Created:         time.Now().Unix(),
		BaseContainerId: container.Id,
		Id:              uuid.NewV4().String(),
		Limit:           image.Limit,
		Network: helper.ArrayMap(container.Networks, func(net types.Network) types.KisaraVmNetwork {
			return types.KisaraVmNetwork{
				Name:   net.Name,
				Type:   types.KISARA_VM_NET_TYPE_HOST,
				Subnet: net.Subnet,
			}
		}),
	}

	return vm_instance, nil
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
