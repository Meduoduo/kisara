# Kisara
Kisara is a Docker cluster management tool for CTF/AWD competitions, which is still under development.

[README](./README.md)|[文档](./README-cn.md)

## Open source agreement
Kisara is open source under the MIT license. You can use it for free, but you must retain the original author information and the original license agreement. If you use Kisara for commercial purposes, you must obtain the author's permission. refer to [LICENSE](./LICENSE) for details.

## Why Kisara
Kisara comes from the name of the female protagonist Kisara from the anime "Kiss of the Rose Princess". She is a demon! And I usually name my projects after beautiful girls, so that's it. (lol)

## The origin of Kisara
Because the Irina project needs to decouple Docker management and needs to access the cluster. It is impossible to use Irina's original single-machine Docker. Therefore, Kisara was born. At the same time, Kisara uses Takina to achieve flexible intranet penetration to expose target machine services. Please make sure that the `/conf/takina_client.yaml` file is configured and that the Takina service is started. For more information, please see [https://github.com/Yeuoly/Takina](https://github.com/Yeuoly/Takina).

## Kisara documentation
Kisara is based on the C/S architecture, allowing one S (Server) and multiple C (Clients) to manage a Kisara cluster. The use of Server and Client will be introduced separately below. Note that the environment of Server and Client can be accessed mutually, such as being located on the public network or the intranet, or being internally penetrated.

## Kisara Client
Multiple clients can belong to the same Kisara cluster. The configuration file of a single Kisara Client is as follows. Please ensure that the configuration document is located in the running directory and the relative path is `conf/kisara-conf.toml`.

### Requires
- Docker Daemon Unix Socket Access, make sure KisaraClient can access DockerDaemon through UnixSock
- Docker Swarm Enable, make sure the host is the manager node of the Docker Swarm cluster
- Takina Server Configuration, because Kisara uses Takina as a container port mapping tool, Takina needs to be configured in advance. Just run Takina Server. For more details, please refer to [https://github.com/Yeuoly/Takina](https://github.com/Yeuoly/Takina)
- GO 1.18+

The content is as follows:
```toml
[kisara]
token = "test" # Authentication token used between Server and Client to ensure that the tokens are the same for both Server and Client
mode = "dev" # Operating mode, either dev or prod
dns = "8.8.8.8" # Default DNS server used by the container, recommended to use public servers such as 8.8.8.8 or 114.114.114.114

[kisaraClient]
address = "0.0.0.0" # Must be configured correctly, this is the address that the user provides to the Server to access the Client, ensuring that the Server's network environment can access this address
port = 25570 # Must be configured correctly, this is the port that the user provides to the Server to access the Client, ensuring that the Server's network environment can access this port
network_in = 52428800 # 50Mbps, the inbound bandwidth of the Client, please adjust it based on the Takina configuration
network_out = 52428800 # 50Mbps, the outbound bandwidth of the Client, please adjust it based on the host configuration
max_container = 80 # The maximum number of containers allowed theoretically
db_path = "db/kisara.db" # The temporary database path of Kisara
network_cidrs = "172.[128-255].[0-255].0/24" # The C-class addresses allowed to be applied for, these addresses will be used as the network pool when Kisara applies for a network

[kisaraServer]
address = "0.0.0.0" # The address of the Kisara Server, the Client will connect to this address and act as its Client node to provide services for it
port = 7474 # The port of the Kisara Server, the Client will connect to this port and act as its Client node to provide services for it

[takina]
token = "testtest" # The token of the Takina Server
```

In addition to Kisara's own configuration, Takina configuration also needs to be configured because Takina is used. The configuration file is located at conf/takina_client.yaml.

The content is as follows. As TakinaClient will run as a service in the container, port and other information needs to be configured.

```yaml
server-name: Takina # Takina service name
token: testtest # Token used between TakinaServer and TakinaClient
takina_port: 40002 # The port of the Takina service, do not modify it if using the default port

nodes: # Node list, multiple service nodes can be selected to share the traffic load
  - 
    address: 0.0.0.0 # The address of the Takina Server
    port: 29979 # The port of the Takina Server, can be changed
    token: testtest # The token of the Takina Server, can be changed
```

After the configuration information is completed, KisaraClient can be compiled.

`go build cmd/client/main.go`

Then execute the compiled program to run the Client side.

`./main`

## Kisara Server
Only one Kisara Server is allowed in the same Kisara cluster, and Kisara Server provides Builtin API for accessing the Kisara cluster. It also requires some configuration, but compared to the Client, only some simple configurations are needed.

The basic configuration is still in `conf/kisara-conf.toml`.

```toml
[kisara]
token = "test" # Token used for authentication between Server and Client, ensure that the Token of Server and Client are the same
mode = "dev" # Running mode, dev or prod

[kisaraServer]
address = "159.75.81.96"
port = 7474
```

Then, import the kisara API package in the project that needs to use Kisara, and you can use it. The demo code below is an arbitrarily written `CreateContainer` function. Most data types are custom, as long as they conform to `kisara_types.RequestLaunchContainer`.

```go
import (
    kisara "github.com/Yeuoly/kisara/src/api"
	kisara_types "github.com/Yeuoly/kisara/src/types"
)

type Docker struct{}

func (c *Docker) CreateContainer(
    image *Image, //image
    uid int, // UID of the user who starts the container, can be unspecified
    client_id string,  //client ID, required
    port_protocol string,  //port that needs to be forwarded, format like 80/tcp,22/tcp
    subnet_name string, // name of the subnet to connect the container to
    module string, //module name, optional
    env_mount ...map[string]string // environment variables and mounted paths required
) (*Container, error) {
	resp, err := kisara.LaunchContainer(kisara_types.RequestLaunchContainer{
		ClientID:     client_id, // If client_id is not specified, Kisara will select a node based on its own complex balancing algorithm.
		Image:        image.Name, //name of the image
		UID:          uid, // UID of the user who starts the container, can be unspecified
		PortProtocol: port_protocol, // port that needs to be forwarded
		SubnetName:   subnet_name, // name of the subnet
		Module:       module, // module name, can be unspecified
		EnvMount:     env_mount, // environment variables and mounted paths
	}, time.Duration(time.Second*60))
	if err != nil {
		Warn("[Kisara] start container error: " + err.Error())
		return nil, err
	}
	if resp.Error != "" {
		Warn("[Kisara] start container error: " + resp.Error)
		return nil, errors.New(resp.Error)
	}
	return (*Container)(&resp.Container), nil
}

```

Kisara also supports some event hooks, as shown in the following demo code. The functionality implemented is to automatically create an "irina-train" network when a node connects to the server.

```go
kisara.RegisterOnNodeConnect(func(client_id string, client *kisara_types.Client) {
	Info("[Kisara] node %s connected", client_id)
	// create irina-train network
	// check if network exists
	Info("[Kisara] Initializing network irina-train")
	resp, err := kisara.ListNetwork(kisara_types.RequestListNetwork{
		ClientID: client_id,
	}, time.Duration(time.Second*30))
	if err != nil {
		Warn("[Kisara] list network failed: %s", err.Error())
		return
	}

	if resp.Error != "" {
		Warn("[Kisara] list network failed: %s", resp.Error)
        return
	}

	for _, network := range resp.Networks {
		if network.Name == "irina-train" {
			return
		}
	}

	_, err = kisara.CreateNetwork(kisara_types.RequestCreateNetwork{
		ClientID: client_id,
		Subnet:   irina_train_subnet,
		Name:     "irina-train",
		HostJoin: false,
	}, time.Duration(time.Second*30))

    if err != nil {
		Warn("[Kisara] create network irina-train failed: %s", err.Error())
	}
	Info("[Kisara] Initializing network irina-train finished")
})
```

Details and interfaces will be documented in the future. For now, please refer to the IDE prompts or view the source code for more information.