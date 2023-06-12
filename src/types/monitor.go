package types

type KisaraNetworkMonitorImage struct {
	ImageName string `json:"image_name"`
}

type KisaraNetworkMonitorContainer struct {
	ContainerId string `json:"container_id"`
}

type KisaraNetworkTest struct {
	// The container to be tested
	ContainerId string `json:"container_id"`
	// The container to test
	TestContainerId string `json:"test_container_id"`
	// The test cmd like "python3 test.py $ip", the $ip will be replaced by the container's ip
	Script string `json:"script"`
}

type KisaraNetworkTestSet struct {
	Containers []KisaraNetworkTest `json:"containers"`
}

type KisaraNetworkTestResult struct {
	// The container to be tested
	ContainerId string `json:"container_id"`
	// Result of the test, bytewise
	Result []byte `json:"result"`
}

type KisaraNetworkTestResultSet struct {
	Results []KisaraNetworkTestResult `json:"results"`
}
