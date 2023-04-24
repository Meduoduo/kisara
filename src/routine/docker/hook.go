package docker

import (
	"github.com/Yeuoly/kisara/src/types"
)

type OnContainerLaunch func(*Docker, types.Container)

type OnContainerStop func(*Docker, types.Container)

type OnNetworkCreate func(*Docker, types.Network)

type BeforeNetworkRemove func(*Docker, types.Network)

type OnNetworkRemove func(*Docker, types.Network)

type OnDockerDaemonStart func(*Docker, []types.Network)

var (
	// OnContainerLaunchHook is called when a container is launched
	onContainerLaunchHook []OnContainerLaunch
	// OnContainerStopHook is called when a container is stopped
	onContainerStopHook []OnContainerStop
	// OnNetworkCreateHook is called when a network is created
	onNetworkCreateHook []OnNetworkCreate
	// BeforeNetworkRemoveHook is called before a network is removed
	beforeNetworkRemoveHook []BeforeNetworkRemove
	// OnNetworkRemoveHook is called when a network is removed
	onNetworkRemoveHook []OnNetworkRemove

	// OnDockerDaemonStartHook is called when the docker daemon is started
	onDockerDaemonStartHook []OnDockerDaemonStart
)

// AddOnContainerLaunchHook adds a hook to the OnContainerLaunchHook list
func AddOnContainerLaunchHook(hook OnContainerLaunch) {
	onContainerLaunchHook = append(onContainerLaunchHook, hook)
}

// AddOnContainerStopHook adds a hook to the OnContainerStopHook list
func AddOnContainerStopHook(hook OnContainerStop) {
	onContainerStopHook = append(onContainerStopHook, hook)
}

// AddOnNetworkCreateHook adds a hook to the OnNetworkCreateHook list
func AddOnNetworkCreateHook(hook OnNetworkCreate) {
	onNetworkCreateHook = append(onNetworkCreateHook, hook)
}

// AddBeforeNetworkRemoveHook adds a hook to the BeforeNetworkRemoveHook list
func AddBeforeNetworkRemoveHook(hook BeforeNetworkRemove) {
	beforeNetworkRemoveHook = append(beforeNetworkRemoveHook, hook)
}

// AddOnNetworkRemoveHook adds a hook to the OnNetworkRemoveHook list
func AddOnNetworkRemoveHook(hook OnNetworkRemove) {
	onNetworkRemoveHook = append(onNetworkRemoveHook, hook)
}

// AddOnDockerDaemonStartHook adds a hook to the OnDockerDaemonStartHook list
func AddOnDockerDaemonStartHook(hook OnDockerDaemonStart) {
	onDockerDaemonStartHook = append(onDockerDaemonStartHook, hook)
}

// callOnContainerLaunchHooks calls all hooks in the OnContainerLaunchHook list
func callOnContainerLaunchHooks(c *Docker, container types.Container) {
	for _, hook := range onContainerLaunchHook {
		hook(c, container)
	}
}

// callOnContainerStopHooks calls all hooks in the OnContainerStopHook list
func callOnContainerStopHooks(c *Docker, container types.Container) {
	for _, hook := range onContainerStopHook {
		hook(c, container)
	}
}

// callOnNetworkCreateHooks calls all hooks in the OnNetworkCreateHook list
func callOnNetworkCreateHooks(c *Docker, network types.Network) {
	for _, hook := range onNetworkCreateHook {
		hook(c, network)
	}
}

// callBeforeNetworkRemoveHooks calls all hooks in the BeforeNetworkRemoveHook list
func callBeforeNetworkRemoveHooks(c *Docker, network types.Network) {
	for _, hook := range beforeNetworkRemoveHook {
		hook(c, network)
	}
}

// callOnNetworkRemoveHooks calls all hooks in the OnNetworkRemoveHook list
func callOnNetworkRemoveHooks(c *Docker, network types.Network) {
	for _, hook := range onNetworkRemoveHook {
		hook(c, network)
	}
}

// callOnDockerDaemonStartHooks calls all hooks in the OnDockerDaemonStartHook list
func callOnDockerDaemonStartHooks(c *Docker, networks []types.Network) {
	for _, hook := range onDockerDaemonStartHook {
		hook(c, networks)
	}
}
