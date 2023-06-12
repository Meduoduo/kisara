package docker

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Yeuoly/kisara/src/types"
)

/*
	As an imagine, a container should be monitored and we can test it network situation.

	We should use another container to run the test, the only requirement is that the container should be able to access the target container's network.

	But the container should contains all test scripts and tools, considier use Python to write the test scripts.

	to different environment need like pwntools, alf, requests, etc. we should use different container to run the test.
	therefore, a  network could have multiple test container, to init the test container, user can specify different requirements.txt to install different tools.
*/

type KisaraNetworkMonitor interface {
	// Run the test container
	//@param network_name: the network name to run the test container
	//@param context: the context to build the test container, should be a tar file and contains dockerfile etc.
	RunNetworkMonitor(network_name string, context io.Reader, message_callback func(string)) (*types.KisaraNetworkMonitorContainer, error)
	// Stop the test container, and automatically remove it
	StopNetworkMonitor(container *types.KisaraNetworkMonitorContainer) error
	// Run the test script, support multiple containers
	RunNetworkMonitorScript(containers *types.KisaraNetworkTestSet) (*types.KisaraNetworkTestResultSet, error)
	// generate test image name
	generateTestImageName(network_name string) string
}

var (
	network_monitor     map[string][]string = make(map[string][]string)
	network_monitor_mux sync.Mutex
)

func registerMonitorToNetwork(network_id string, test_container_id string) {
	network_monitor_mux.Lock()
	defer network_monitor_mux.Unlock()

	// check if the network exists
	if _, ok := network_monitor[network_id]; !ok {
		network_monitor[network_id] = make([]string, 0)
	}

	// check if the container has been registered
	for _, container_id := range network_monitor[network_id] {
		if container_id == test_container_id {
			return
		}
	}

	// register the container
	network_monitor[network_id] = append(network_monitor[network_id], test_container_id)
}

func unregisterMonitorFromNetwork(network_id string, test_container_id string) {
	network_monitor_mux.Lock()
	defer network_monitor_mux.Unlock()

	// check if the network exists
	if _, ok := network_monitor[network_id]; !ok {
		return
	}

	// check if the container has been registered
	for index, container_id := range network_monitor[network_id] {
		if container_id == test_container_id {
			network_monitor[network_id] = append(network_monitor[network_id][:index], network_monitor[network_id][index+1:]...)
			return
		}
	}
}

// return the network id of the test container
func getMonitorContainerNetwork(monitor_container_id string) (string, error) {
	network_monitor_mux.Lock()
	defer network_monitor_mux.Unlock()

	for network_id, container_ids := range network_monitor {
		for _, container_id := range container_ids {
			if container_id == monitor_container_id {
				return network_id, nil
			}
		}
	}

	return "", fmt.Errorf("monitor container: %s not found", monitor_container_id)
}

func (c *Docker) RunNetworkMonitor(network_name string, context io.Reader, message_callback func(string)) (*types.KisaraNetworkMonitorContainer, error) {
	// build the image
	image_name := c.generateTestImageName(network_name)

	// check if the image exists
	if c.CheckImageExist(image_name) {
		// remove the image
		err := c.DeleteImage(image_name)
		if err != nil {
			return nil, fmt.Errorf("remove image: %s, %s", image_name, err.Error())
		}
	}

	finished_chan := make(chan struct{})
	var fault_error error

	err := c.BuildImage(
		context,
		image_name,
		func(message string) {
			message_callback(fmt.Sprintf("build image: %s, %s", image_name, message))
		},
		func(fault string) {
			fault_error = fmt.Errorf("build image: %s, %s", image_name, fault)
			message_callback(fmt.Sprintf("fault of building image: %s, %s", image_name, fault))
		},
		finished_chan,
	)

	if err != nil {
		return nil, err
	}

	// wait for the build finished
	<-finished_chan

	if fault_error != nil {
		return nil, fault_error
	}

	network, err := c.GetNetworkByName(network_name)
	if err != nil {
		return nil, err
	}

	// run the container
	container, err := c.LaunchContainer(image_name, 0, "", network_name, "checker")
	if err != nil {
		return nil, err
	}

	// register the container to the network
	registerMonitorToNetwork(network.Id, container.Id)

	kisara_network_monitor_container := &types.KisaraNetworkMonitorContainer{
		ContainerId: container.Id,
	}

	return kisara_network_monitor_container, nil
}

func (c *Docker) StopNetworkMonitor(container *types.KisaraNetworkMonitorContainer) error {
	container_id := container.ContainerId

	network, err := c.GetContainerNetwork(container_id)
	if err != nil {
		return err
	}

	// stop the container
	err = c.StopContainer(container_id)
	if err != nil {
		return err
	}

	// unregister the container from the network
	for _, container_network := range network.Networks {
		unregisterMonitorFromNetwork(container_network.Network.Id, container_id)
	}

	// remove the image
	err = c.DeleteImage(container_id)
	if err != nil {
		return err
	}

	return nil
}

/*
Run the test script, support multiple containers
result and error should be all considered beacuse of multiple tests
*/
func (c *Docker) RunNetworkMonitorScript(containers *types.KisaraNetworkTestSet) (*types.KisaraNetworkTestResultSet, error) {
	// run the test script
	results := &types.KisaraNetworkTestResultSet{}
	var errs []error

	var wg sync.WaitGroup
	for _, container := range containers.Containers {
		// get the network of the container
		container_networks, err := c.GetContainerNetwork(container.ContainerId)
		if err != nil {
			return nil, err
		}

		ip := ""
		test_network_id, err := getMonitorContainerNetwork(container.TestContainerId)
		if err != nil {
			continue
		}

		for _, container_network := range container_networks.Networks {
			if container_network.Network.Id == test_network_id {
				ip = container_network.Ip
				break
			}
		}

		if ip == "" {
			continue
		}

		wg.Add(1)
		go func(container_id string, cmd string) {
			// run the test script
			cmd = strings.Replace(cmd, "$ip", ip, -1)
			result, err := c.ExecWarp(container_id, cmd, time.Second*10)
			if err != nil {
				errs = append(errs, err)
			} else {
				results.Results = append(results.Results, types.KisaraNetworkTestResult{
					ContainerId: container_id,
					Result:      result,
				})
			}
			wg.Done()
		}(container.ContainerId, container.Script)
	}

	wg.Wait()
	return results, errors.Join(errs...)
}

func (c *Docker) generateTestImageName(network_name string) string {
	return fmt.Sprintf("kisara_network_monitor_%s", network_name)
}
