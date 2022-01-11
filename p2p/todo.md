# todo
1. replay没有实现 
2. lamport 时间用于回滚纠错？
3. 使用level持久化数据库 - 完成


## must do
- Gossip: memberlist
- mDNS
- gRPC
- BlockChain: Tendermint
- Lamport time O’Clock
- snapshot
- replay


## webAPI

- /put 写入数据
  - key 需要保存的键
  - val 需要保存的值
  - 返回 文本
- /del 删除数据
  - key
  - 返回 文本
- /get 查看数据
  - key
  - 返回 文本
- /kv 查看全部数据
  - 返回 列表json
- /join 加入节点
  - member exp: 192.168.3.3:12345
  - 返回 文本
- /info 显示信息
  - 返回 json
- /start 开启节点
  - local_name 本机名称，默认用机器名称
  - cluster_name 自动发现的集群名称，默认
 mycluster
  - port 端口号，默认随机闲置
  - member 同上
  - 返回文本
- /stop 停止节点
  - 返回 文本