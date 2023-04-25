# Kisara
Kisara 是一个用于CTF/AWD等竞赛的Docker集群管理工具，现在尚处于研发阶段

[README](./README.md)|[文档](./README-cn.md)

## 为什么叫Kisara
来自于我很喜欢的动画《契约之吻》的女主Kisara，她是一个恶魔！而我对项目的命名基本都是美少女，所以就这样了（雾

## Kisara的由来
因为Irina（伊蕾娜）项目中需要对Docker管理进行解耦，同时需要接入集群，使用Irina原本的单机Docker是无法实现的，因此有了Kisara，
同时Kisara使用了Takina（泷奈）来实现灵活的内网穿透从而暴露靶机服务，请确保配置了 `/conf/takina_client.yaml` 文件，并且确保Takina服务启动，详细信息请查看 [https://github.com/Yeuoly/Takina](https://github.com/Yeuoly/Takina)

## Kisara文档
Kisara基于C/S架构，允许一个S（Server）和多个C（Client）管理一个Kisara集群，下面将分别介绍Server和Client的使用，注意，Server和Client的环境可以相互访问，如同时位于公网或同时处于内网，或者对其做内网穿透处理

## Kisara Client
同一个Kisara集群可以拥有多个Client，单个Kisara Client的配置文件如下，请确保配置文档最终位于运行目录下，相对路径为 `conf/kisara-conf.toml`

### Requires
 - Docker Daemon Unix Socket Access，确保KisaraClient可以通过UnixSock访问DockerDaemon
 - Docker Swarm Enable，确保主机为Docker Swarm集群的Manager节点
 - Takina Server Configuration，由于Kisara使用了Takina作为容器端口映射工具，所以需要提前配置Takina，只需要运行Takina Server即可，详细请看 [https://github.com/Yeuoly/Takina](https://github.com/Yeuoly/Takina)
 - GO 1.18+


内容如下
```toml
[kisara]
token = "test" # 用于Server和Client之间的认证Token，确保Server和Client的Token相同
mode = "dev" # 运行模式，dev或prod
dns = "8.8.8.8" # 容器使用的默认DNS服务器，建议使用8.8.8.8或114.114.114.114等公共服务器

[kisaraClient]
address = "0.0.0.0" # 必须配置正确，用户提供给Server访问Client的地址，确保Server的网络环境可以访问到这个地址
port = 25570 # 必须配置正确，用户提供给Server访问Client的端口，确保Server的网络环境可以访问到这个端口
network_in = 52428800 # 50Mbps，Client入网带宽，请根据Takina配置而定
network_out = 52428800 # 50Mbps，Client出网带宽，请根据主机配置而定
max_container = 80 # 理论允许的最多容器数量
db_path = "db/kisara.db" # Kisara临时数据库路径
network_cidrs = "172.[128-255].[0-255].0/24" # 允许申请的C段地址，这里的地址将作为Kisara申请网络时的网络池

[kisaraServer]
address = "0.0.0.0" # Kisara Server的地址，Client将会连接上这个地址，并作为其Client节点并为其提供服务
port = 7474 # Kisara Server的端口，Client将会连接上这个端口，并作为其Client节点并为其提供服务

[takina]
token = "testtest" # Takina Server的Token
```

除了Kisara自身的配置，由于使用了Takina，因此还需要配置Takina配置，位于 `conf/takina_client.yaml`

内容如下，因为TakinaClient会作为服务运行在容器内，因此需要配置端口等信息

```yaml
server-name: Takina # Takina服务名
token: testtest # 用于TakinaServer和TakinaClient之间的token
takina_port: 40002 # Takina服务的端口，使用默认端口请勿修改

nodes: #节点列表，可以选择多个服务节点来分摊流量压力
  - 
    address: 0.0.0.0 # Takina Server的地址
    port: 29979 # Takina Server的端口，可以更改
    token: testtest # Takina Server的Token，可以更改
```

配置信息完成以后可以编译KisaraClient

`go build cmd/client/main.go`

随后执行编译后的程序即可运行Client端

`./main`

## Kisara Server
同一个Kisara集群只允许拥有一个Server，并且Kisara Server提供Builtin的API用于访问Kisara集群，并且也需要一些配置，不过相比于Client，只需要一些最简单的配置即可

基础配置仍然为 `conf/kisara-conf.toml`
```toml
[kisara]
token = "test" # 用于Server和Client之间的认证Token，确保Server和Client的Token相同
mode = "dev" # 运行模式，dev或prod

[kisaraServer]
address = "159.75.81.96"
port = 7474

```

随后，在需要使用Kisara的项目中引入kisara API包即可，demo代码如下，下面是随意编写的一个CreateContainer函数，数据类型大多数为自定义，只需要符合 `kisara_types.RequestLaunchContainer` 即可

```go
import (
    kisara "github.com/Yeuoly/kisara/src/api"
	kisara_types "github.com/Yeuoly/kisara/src/types"
)

type Docker struct{}

func (c *Docker) CreateContainer(
    image *Image, //镜像
    uid int, // 启动这个容器的uid，可以不指定
    client_id string,  //指定
    port_protocol string,  //需要转发的端口，格式如 80/tcp,22/tcp
    subnet_name string, // 容器连接的子网名称
    module string, //模块名，
    env_mount ...map[string]string // 需要
) (*Container, error) {
	resp, err := kisara.LaunchContainer(kisara_types.RequestLaunchContainer{
		ClientID:     client_id, // 如果不指定client_id，Kisara将会根据自身复杂均衡算法选择节点
		Image:        image.Name, //镜像名
		UID:          uid, // 用户ID，可以不指定
		PortProtocol: port_protocol, // 需要转发的端口
		SubnetName:   subnet_name, // 子网名
		Module:       module, // 模块，可以不指定
		EnvMount:     env_mount, // 环境变量和挂载路径
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

同时Kisara支持一些事件Hook，demo如下，实现功能为当有节点连接上Server的时候自动创建irina-train网络
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

详细信息和接口会在未来完成文档，现在请根据IDE提示或者查看源码