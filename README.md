# IM消息中台
## 确定业务场景
面对的业务场景是什么： 直播聊天室，企业级im, 单聊， 群聊，多聊（聊天室内）
### 读放大，写放大
什么业务场景会产生读写放大，解决方案
## 消息的储存结构 会话领域
消息的优先级， 消息的撤回， 消息ID, 会话ID
MsgID 如何生成
怎么储存消息， 消息的结构是什么 **会话链**， **混合链**。
使用什么数据库， 如何分布式化
### 如何GC
GC的对象是什么
不同的业务场景GC方式不同。 如直播聊天室，当直播结束就可以直接回收。企业IM需要永久储存。微信需要保留一段时间内的数据， 引出在客户端还是服务端的问题，微信保持在客户端。
#### 储存数据结构
**环形缓冲区** 分布式 实现环形缓冲区域，头删除，尾插
**merge list**储存在分布式KV中 读Big table论文 重点了解
用户链(权限管理)：key = sessionID + userID
会话链: key = sessionID
key => ----zset---- size = 5000 后分裂 插入meta info    
## 储存服务
**最终一致性** **因果一致性**
什么是因果一致性，checkpoint
自研分布式储存系统，直接在原生语 义上支持ringbuffer和merge list的api
meta server: 管理分片以及生命周期
## 消息领域中台化 
作为一个单独的中台，为各种更上层服务服务。如在直播，聊天APP,短视频中作为消息系统
## 用户领域中台化
如在直播，聊天APP,短视频中作为用  户认证1登陆xi togn
## Push server
负载均衡 聚合 如同一个聊天时的用户聚合在一台 push server上
背景介绍
业务需求
设计目标
基本解法
技术挑战
实现细节
## 实现细节
### IM SDK
- 消息补偿 
- 消息回退 /推拉模式切换 降级为拉模式 升级为推模式
- 消息写入 write server
- 消息读入 read server
### IM http server
- web 接口搭建
- 支持一个http接口， 部署ng网关 why /logic： 处理域名问题
- 登陆态检查/限流/消息类型判断等业务逻辑
- 调用imskd消息补偿接口
- 做读取限流处理
### router server改造
router/table.go
- 支持 群聊/聊天室 sessionID路由注册
- 支持 sessionID到userID; userID到endpoint(gateway endpoint什么服务的端点)的双跳查询
- 在实现上可以通过心跳不断感知客户端的状态变化来优化路由表，做出正确的路由策略
- 一些思考在plato1中该server管理的是did 到gataway endpoint ip:port的映射。在plato2中应该管理哪些映射。
- state server 与 gateway server可以作为两个进程运行在一台物理机上。也可以分布式部署，各自作为一个集群存在。plato2中应是偏向后者
### ipconfig
- 在plato1中ipconfig作为load balancer仅仅对gateway server做了负载均衡。在plato2中要考虑对state server做负载均衡
- 支持 直播(聊天室)/群聊 按直播间聚合分配策略 （直播场景下sessionID 与userID一一对应，故尽可能将同一直播室中的用户分在同一个state/gateway server）
- 超大群打散均衡策略，大群聊分片
### push(state) server
- 支持时间轮触发通过注册的sessionID集合拉取自己负责的task(什么task，那些task)
- 周期性拉取消息写入到本地缓存中并且通过时间戳标记是否拉取
- 支持时间轮触发push任务，从本地拉取没有被推送的消息，均匀将一个房间的用户分发到300个channel上
- 通过sessionID从router server中查询具体im gateway endpoint地址（在plato1中是connID的到gateway endpoint）
- 针对每个im gateway打包一个batch msg结构一次给gateway, 有uid的list以及一个群聊当前的msgData
- 支持c2c rpc推送消息，userID反查gateway endpoint推送出去
- 支持 群聊/c2c/等因果一致性消息last msg定时器逻辑，支持ack在msgID匹配时才取消定时
- 支持超大群聊活跃时协议回退rpc（从MQ回退到RPC）
### Writer server
- 按消息类型进行最终一致性或因果一致性存储
- 支持按sessionID维度做到写入频控识别，高热群聊发送协议回退消息
- 发送消息到MQ, 由消息同步组件消费，跨数据中心同步
- 对 单聊/小群聊走c2c rpc直接push消息
- 缓存meta server的路由信息 /什么
- 当请求下游分片路由失败时，回原路由信息 /什么
- 对于因果一致性存储，按消息类型写入用户链和会话链接
### meta server
-  分片路由信息管理（CURD）
-  index server全生命周期管理（状态机流转）
-  心跳检查，检查信息时同步index server角色与状态
-  选主，切主
-  实现新老配置融合分片
-  感知follower热点后，自动将buckup 节点升级为learner节点，热点扩容与缩容逻辑
### index server
- 实现wal基本结构，mmap机制/checkpoint/按时间和数量切文件流
- leader角色支持消息写入wal,当自己不是leader后拒绝写入
- sessionID不在自己分片上拒绝写入
- Follower拉去wal日志构建ringbuffer数据，对外支持range本地消息(iterator)
- Learner拉去wal 记录lag不达到配置阈值不支持读取
- buckup角色不提供任务服务，作为init状态
### reader server
- 通过meta server读取分片信息，决定sessionID hash到哪个分片
- 负载均衡到该分片的某个follower上查询本地对应的ringbuffer
- 如果是因果一致性存储，对 会话链和用户链读取后按msgID排序后合并，指定preMsgID构建消息链
### corekv / tikv
- 搭建tikv || corekv分布式集群
- 构建 merge list:
- ring buffer 结构支持
- 根据消息类型，写入消息到会话链和用户链
