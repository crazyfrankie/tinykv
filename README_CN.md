# TinyKV 课程

TinyKV 课程通过 Raft 共识算法构建一个键值存储系统。它的灵感来源于 [MIT 6.824](https://pdos.csail.mit.edu/6.824/) 和 [TiKV 项目](https://github.com/tikv/tikv)。

完成本课程后，您将具备实现一个水平可扩展、高可用、支持分布式事务的键值存储服务的知识。同时，您将更好地理解 TiKV 的架构和实现。

## 课程架构

整个项目在开始时是一个键值服务器和调度器服务器的框架代码 - 您需要逐步完成核心逻辑：

* [独立 KV](https://github.com/crazyfrankie/tinykv/blob/master/doc/project1-StandaloneKV_CN.md)
  * 实现一个独立的存储引擎。
  * 实现原始键值服务处理程序。
* [Raft KV](https://github.com/crazyfrankie/tinykv/blob/master/doc/project2-RaftKV_CN.md)
  * 实现基本的 Raft 算法。
  * 在 Raft 之上构建容错 KV 服务器。
  * 添加 Raft 日志垃圾回收和快照支持。
* [多 Raft KV](https://github.com/crazyfrankie/tinykv/blob/master/doc/project3-MultiRaftKV_CN.md)
  * 为 Raft 算法实现成员变更和领导者变更。
  * 在 Raft 存储上实现配置变更和区域分裂。
  * 实现基本的调度器。
* [事务](https://github.com/crazyfrankie/tinykv/blob/master/doc/project4-Transaction_CN.md)
  * 实现多版本并发控制层。
  * 实现 `KvGet`、`KvPrewrite` 和 `KvCommit` 请求的处理程序。
  * 实现 `KvScan`、`KvCheckTxnStatus`、`KvBatchRollback` 和 `KvResolveLock` 请求的处理程序。

## 代码结构

![概览](https://github.com/crazyfrankie/tinykv/blob/master/doc/imgs/overview.png)

类似于 TiDB + TiKV + PD 分离存储和计算的架构，TinyKV 只专注于分布式数据库系统的存储层。如果您对 SQL 层也感兴趣，请参阅 [TinySQL](https://github.com/tidb-incubator/tinysql)。此外，还有一个名为 TinyScheduler 的组件，作为整个 TinyKV 集群的中央控制，收集来自 TinyKV 心跳的信息。之后，TinyScheduler 可以生成调度任务并将任务分发给 TinyKV 实例。所有实例通过 RPC 进行通信。
整个项目组织为以下目录：

* `kv` 包含键值存储的实现。
* `raft` 包含 Raft 共识算法的实现。
* `scheduler` 包含 TinyScheduler 的实现，负责管理 TinyKV 节点和生成时间戳。
* `proto` 包含所有节点和进程之间通过 gRPC 上的 Protocol Buffers 进行的通信。该包包含 TinyKV 使用的协议定义和生成的 Go 代码。
* `log` 包含基于级别输出日志的实用工具。

## 阅读列表

我们提供了一个[阅读列表](https://github.com/crazyfrankie/tinykv/blob/master/doc/reading_list_CN.md)，涵盖分布式存储系统的知识。虽然并非所有内容都与本课程高度相关，但它们可以帮助您构建该领域的知识体系。

此外，我们鼓励您阅读 TiKV 和 PD 设计的概述，以对您将要构建的内容有一个总体印象：

* TiKV，数据存储的设计（[英文](https://en.pingcap.com/blog/tidb-internal-data-storage)，[中文](https://pingcap.com/zh/blog/tidb-internal-1)）。
* PD，调度的设计（[英文](https://en.pingcap.com/blog/tidb-internal-scheduling)，[中文](https://pingcap.com/zh/blog/tidb-internal-3)）。

## 从源代码构建 TinyKV

### 前提条件

* `git`：TinyKV 的源代码托管在 GitHub 上作为 git 仓库。要使用 git 仓库，请[安装 `git`](https://git-scm.com/downloads)。
* `go`：TinyKV 是一个 Go 项目。要从源代码构建 TinyKV，请[安装 `go`](https://golang.org/doc/install)，版本需大于或等于 1.13。

### 克隆

将源代码克隆到您的开发机器上。

```bash
git clone https://github.com/tidb-incubator/tinykv.git
```

### 构建

从源代码构建 TinyKV。

```bash
cd tinykv
make
```

这会将 `tinykv-server` 和 `tinyscheduler-server` 的二进制文件构建到 `bin` 目录中。
## 与 TinySQL 一起运行 TinyKV

1. 按照[其文档](https://github.com/tidb-incubator/tinysql#deploy)获取 `tinysql-server`。
2. 将 `tinyscheduler-server`、`tinykv-server` 和 `tinysql-server` 的二进制文件放入同一个目录中。
3. 在二进制文件目录下，运行以下命令：

```bash
mkdir -p data
./tinyscheduler-server
./tinykv-server -path=data
./tinysql-server --store=tikv --path="127.0.0.1:2379"
```

现在您可以使用官方 MySQL 客户端连接到数据库：

```bash
mysql -u root -h 127.0.0.1 -P 4000
```

## 自动评分和认证

自 2022 年 6 月起，我们开始使用 [github classroom](https://github.com/talent-plan/tinysql/blob/course/classroom.md) 接收实验并及时提供自动评分。GitHub classroom 邀请链接是 https://classroom.github.com/a/cdlNNrFU。讨论微信/Slack 群组和通过课程后的认证在 [tinyKV 学习课程](https://talentplan.edu.pingcap.com/catalog/info/id:263) 中提供。

自动评分是一个可以自动运行测试用例并及时给出反馈的工作流。然而，GitHub classroom 有一些限制，为了使 golang 工作并在我们的自托管机器上运行，**您需要覆盖 Github classroom 生成的工作流并提交它**。

```sh
cp scripts/classroom.yml .github/workflows/classroom.yml
git add .github
git commit -m"update github classroom workflow"
```

## 贡献

非常感谢任何反馈和贡献。如果您想参与开发，请查看 [issues](https://github.com/tidb-incubator/tinykv/issues)。
