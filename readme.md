``` raw
███╗   ██╗███████╗ ██████╗ ███╗   ███╗███████╗ ██████╗  █████╗   
████╗  ██║██╔════╝██╔═══██╗████╗ ████║██╔════╝██╔════╝ ██╔══██╗  
██╔██╗ ██║█████╗  ██║   ██║██╔████╔██║█████╗  ██║  ███╗███████║  
██║╚██╗██║██╔══╝  ██║   ██║██║╚██╔╝██║██╔══╝  ██║   ██║██╔══██║  
██║ ╚████║███████╗╚██████╔╝██║ ╚═╝ ██║███████╗╚██████╔╝██║  ██║  
╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝  
``` 

## neomega
### 目的
`neomega` 提供了与 `Minecraft` 相关的工具、概念抽象和 `API` ，但是 仅仅包括 `Minecraft` 相关的 `API` 。
考虑到使用该库的项目可能包含了一些复杂的，很可能出现错误的业务逻辑，因此 `neomega` 支持以远程附加形式工作，即:
  - 核心部分，即机器人本体，尽可能精简代码和业务逻辑，以期长时间稳定运行，同时开放数据接口
  - 业务部分，即运行复杂的，可能出现错误逻辑的部分，通过上述数据接口和核心建立连接，避免因为业务本身的错误拖累核心部分正常工作
### 覆盖范围
- `Netease Minecraft` 的登陆,验证服务器相关，以及后处理
- 机器人基本 `API` 及 `Minecraft` 相关基本 `API` 的封装
  - 支持对上述 `API` 的数据和状态的保存、更新和网络传输
  - 一对远程访问上述 `API` 和数据的 `Access Point` 和 `End Point`
  - 机器人动作，例如向箱子放置东西，丢出物品，附魔，重命名等等
  - 所有支持的接口描述均位于 neomega/*.go
- 方块转换和描述: neomega/blocks
- c 接口 neomega/c_api
- 区块转换和描述 neomega/mirror
- 一些基本的例子可以在 neomega/entries 下找到

### 特别说明:
- 此仓库不包含机器人登陆时信息的获取
- 对于一些和机器人相关的数据包，例如 ChangeDimension, Respawn, 其信息的记录位于 uqHolder, 但是响应应该位于 botAction  
- 为了减轻通信量，一些数据包被过滤或者不转发，包括:  
  - packet.MovePlayer: pk.EntityRuntimeID != botRuntimeID 时忽略
  - packet.IDMoveActorDelta: pk.EntityRuntimeID != botRuntimeID 时忽略
  - packet.IDSetActorData: pk.EntityRuntimeID != botRuntimeID 时忽略
  - packet.IDSetActorMotion: pk.EntityRuntimeID != botRuntimeID 时忽略

### 相关库
- gophertunnel: 
https://github.com/Sandertv/gophertunnel  
gophertunnel 是一个国际版 MC 相关库，其实现了非常多数据包的序列化/反序列化 & 连接等重要协议，但是因为网易 MC 和 国际版 MC 不同，我们复制了这个库并对其修改以使之可以与网易 MC 相兼容

 
- Tnze/go-mc   
https://github.com/Tnze/go-mc  
尽管不是很显眼，但是 snbt/nbt 等实现使用的是 Tnze/go-mc

 
- PhoenixBuilder  
https://github.com/LNSSPsd/PhoenixBuilder  
PhoenixBuilder主要用于网易版MC的建筑导入导出，几何体构建等。neomega 前身由 PhoenixBuilder 分叉出来

## nodes
### 目的
`nodes` 旨在将多个分布的节点和进程通过网络或IPC连接起来(目前使用 zmq 库)，提供统一易用的API，并从 API 层面尽量屏蔽不同节点之间的差别  

### 覆盖范围
`nodes` 包含一个主节点 (master) 和多个从节点 (slave)  
`nodes` 包含五类主要 API：消息发布、远程调用 (RPC)、数据读写、节点 Tag 信息、 资源锁
其中，每一类 API 的动机如下:  
- 消息发布: 任意节点均可发布信息，且任意订阅节点均可收到信息，多对多，信息的发布和订阅按 topic 聚合
- 远程调用: 任意节点均可暴露函数和API，任意节点均可按 api name 远程调用此 API，单对单
- 数据读写: 任意节点均可读写整个网络中特定数据，表现的就好像是一个本地的 map
- 节点 Tag 信息: 表示网络中具有具备了某些能力的节点，一个节点可能有多个 tag, 例如 (access point, player command sender, advance bot control)
- 资源锁: 任意节点均可争用网络中的某个锁（一定时间），并可归还该锁