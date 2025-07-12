# Project2 RaftKV

Raft 是一种旨在易于理解的共识算法。你可以在 [Raft 网站](https://raft.github.io/)上阅读关于 Raft 本身的资料，还有一个 Raft 的交互式可视化工具，以及其他资源，包括 [Raft 的扩展论文](https://raft.github.io/raft.pdf)。

在这个项目中，你将基于 Raft 实现一个高可用的 KV 服务器，这不仅需要你实现 Raft 算法，还需要实际应用它，并为你带来更多挑战，例如使用 `badger` 管理 Raft 的持久化状态，为快照消息添加流控制等。

该项目有 3 个部分需要你完成，包括：

- 实现基本的 Raft 算法
- 在 Raft 之上构建一个容错的 KV 服务器
- 添加对 raftlog GC 和快照的支持

## Part A

### 代码

在这一部分，你将实现基本的 Raft 算法。你需要实现的代码在 `raft/` 目录下。在 `raft/` 内部，有一些骨架代码和测试用例等着你。你将要在这里实现的 Raft 算法与上层应用程序有一个精心设计的接口。此外，它使用一个逻辑时钟（这里称为 tick）来衡量选举和心跳超时，而不是物理时钟。也就是说，不要在 Raft 模块本身设置定时器，上层应用程序负责通过调用 `RawNode.Tick()` 来推进逻辑时钟。除此之外，消息的发送和接收以及其他事情都是异步处理的，也由上层应用程序决定何时实际执行这些操作（详见下文）。例如，Raft 不会阻塞等待任何请求消息的响应。

在实现之前，请先查看这部分的提示。此外，你应该大致看一下 `proto/proto/eraftpb.proto` 这个 proto 文件。Raft 发送和接收的消息以及相关的结构体都在那里定义，你将在实现中使用它们。请注意，与 Raft 论文不同，它将心跳和附加条目（AppendEntries）分成不同的消息，以使逻辑更清晰。

这部分可以分为 3 个步骤，包括：

- Leader 选举
- 日志复制
- Raw node 接口

### 实现 Raft 算法

`raft/raft.go` 中的 `raft.Raft` 提供了 Raft 算法的核心，包括消息处理、驱动逻辑时钟等。更多实现指南，请查看 `raft/doc.go`，其中包含了设计概述以及这些 `MessageTypes` 的职责。

#### Leader 选举

要实现 Leader 选举，你可能想从 `raft.Raft.tick()` 开始，它用于将内部逻辑时钟推进一个 tick，从而驱动选举超时或心跳超时。你现在不需要关心消息的发送和接收逻辑。如果你需要发送消息，只需将其推送到 `raft.Raft.msgs` 中，所有 Raft 收到的消息都将传递给 `raft.Raft.Step()`。测���代码将从 `raft.Raft.msgs` 获取消息，并通过 `raft.Raft.Step()` 传递响应消息。`raft.Raft.Step()` 是消息处理的入口，你应该处理像 `MsgRequestVote`、`MsgHeartbeat` 及其响应这样的消息。并且请实现测试桩函数并让它们被正确调用，比如 `raft.Raft.becomeXXX`，它用于在 Raft 角色改变时更新 Raft 的内部状态。

你可以运行 `make project2aa` 来测试实现，并在本部分末尾看到一些提示。

#### 日志复制

要实现日志复制，你可能想从在发送方和接收方处理 `MsgAppend` 和 `MsgAppendResponse` 开始。查看 `raft/log.go` 中的 `raft.RaftLog`，这是一个帮助你管理 Raft 日志的辅助结构体，在这里你还需要通过 `raft/storage.go` 中定义的 `Storage` 接口与上层应用程序交互，以获取持久化的数据，如日志条目和快照。

你可以运行 `make project2ab` 来测试实现，并在本部分末尾看到一些提示。

### 实现 Raw node 接口

`raft/rawnode.go` 中的 `raft.RawNode` 是我们与上层应用程序交互的接口，`raft.RawNode` 包含 `raft.Raft` 并提供一些包装函数，如 `RawNode.Tick()` 和 `RawNode.Step()`。它还提供 `RawNode.Propose()` 让上层应用程序提议新的 Raft 日志。

另一个重要的结构体 `Ready` 也在这里定义。在处理消息或推��逻辑时钟时，`raft.Raft` 可能需要与上层应用程序交互，例如：

- 发送消息给其他 peer
- 将日志条目保存到稳定存储
- 将诸如任期、提交索引和投票等硬状态保存到稳定存储
- 将已提交的日志条目应用到状态机
- 等等

但这些交互不会立即发生，而是被封装在 `Ready` 中，并通过 `RawNode.Ready()` 返回给上层应用程序。由上层应用程序决定何时调用 `RawNode.Ready()` 并处理它。处理完返回的 `Ready` 后，上层应用程序还需要调用一些函数，如 `RawNode.Advance()` 来更新 `raft.Raft` 的内部状态，如应用的索引、稳定的日志索引等。

你可以运行 `make project2ac` 来测试实现，并运行 `make project2a` 来测试整个 A 部分。

> 提示：
>
> - 将你需要的任何状态添加到 `raft.Raft`、`raft.RaftLog`、`raft.RawNode` 和 `eraftpb.proto` 中的消息中。
> - 测试假设第一次启动 Raft 时任期为 0。
> - 测试假设新当选的 Leader 应该在其任期内附加一个 noop 条目。
> - 测试假设一旦 Leader 推进其提交索引，它将通过 `MessageType_MsgAppend` 消息广播提交索引。
> - 测试没有为本地消息 `MessageType_MsgHup`、`MessageType_MsgBeat` 和 `MessageType_MsgPropose` 设置任期。
> - Leader 和非 Leader 附加日志条目的方式有很大不同，有不同的来源、检查和处理，请小心。
> - 不要忘记不同 peer 之间的选举超时应该是不同的。
> - `rawnode.go` 中的一些包装函数可以用 `raft.Step(local message)` 来实现。
> - 启动一个新的 Raft 时，从 `Storage` 获取最后稳定的状态来初始化 `raft.Raft` 和 `raft.RaftLog`。

## Part B

在这一部分，你将使用在 A 部分中实现的 Raft 模块构建一个容错的键值存储服务。你的键值服务将是一个复制状态机，由多个使用 Raft 进行复制的键值服务器组成。只要大多数服务器存活并且可以通信，你的键值服务就应该继续处理客户端请求，尽管存在其他故障或网络分区。

在 project1 中，你已经实现了一个独立的 KV 服务器，所以你应该已经熟悉 KV 服务器的 API 和 `Storage` 接口。

在介绍代码之前，你需要先理解三个术语：`Store`、`Peer` 和 `Region`，它们在 `proto/proto/metapb.proto` 中定义。

- Store 代表一个 tinykv-server 实例
- Peer 代表一个在 Store 上运行的 Raft 节点
- Region 是 Peer 的集合，也称为 Raft 组

![region](imgs/region.png)

为简单起见，在 project2 中，一个 Store 上只有一个 Peer，一个集群中只有一个 Region。所以你现在不需要考虑 Region 的范围。多个 Region 将在 project3 中进一步介绍。

### 代码

首先，你应该看一下 `kv/storage/raft_storage/raft_server.go` 中的 `RaftStorage`，它也实现了 `Storage` 接口。与直接向底层引擎写入或读取的 `StandaloneStorage` 不同，它首先将每个写和读请求发送给 Raft，然后在 Raft 提交请求后，再对底层引擎进行实际的写和读。通过这种方式，它可以保持多个 Store 之间的一致性。

`RaftStorage` 创建一个 `Raftstore` 来驱动 Raft。当调用 `Reader` 或 `Write` 函数时，它实际上通过通道（`raftWorker` 的 `raftCh` 通道）向 raftstore 发送一个在 `proto/proto/raft_cmdpb.proto` 中定义的 `RaftCmdRequest`，该请求带有四种基本命令类型（Get/Put/Delete/Snap），并在 Raft 提交和应用该命令后返回响应。`Reader` 和 `Write` 函数的 `kvrpc.Context` 参数现在很有用，它从客户端的角度携带了 Region 信息，并作为 `RaftCmdRequest` 的头部传递。这些信息可能不正确或已过时，所以 raftstore 需要检查它们并决定是否提议该请求。

然后，就到了 TinyKV 的核心——raftstore。结构有点复杂，阅读 TiKV 的参考资料以更好地理解其设计：

- <https://pingcap.com/blog-cn/the-design-and-implementation-of-multi-raft/#raftstore> (中文版)
- <https://pingcap.com/blog/design-and-implementation-of-multi-raft/#raftstore> (英文版)

raftstore 的入口是 `Raftstore`，见 `kv/raftstore/raftstore.go`。它启动一些 worker 来异步处理特定任务，其中大部分现在还用不到，所以你可以忽略它们。你只需要关注 `raftWorker`。(kv/raftstore/raft_worker.go)

整个过程分为两部分：raft worker 轮询 `raftCh` 以获取消息，包括驱动 Raft 模块的基本 tick 和要作为 Raft 条目提议的 Raft 命令；它从 Raft 模块获取并处理 ready，包括发送 raft 消息、持久化状态、将已提交的条目应用到状态机。一旦应用，就向客户端返回响应。

### 实现 peer storage

Peer storage 是你在 A 部分通过 `Storage` 接口与之交互的东西，但除了 raft 日志，peer storage 还管理其他持久化的元数据，这对于在重启后恢复一致的状态机非常重要。此外，在 `proto/proto/raft_serverpb.proto` 中定义了三个重要的状态：

- RaftLocalState: 用于存储当前 Raft 的 HardState 和最后的日志索引。
- RaftApplyState: 用于存储 Raft 应用的最后日志索引和一些被截断的日志信息。
- RegionLocalState: 用于存储 Region 信息和此 Store 上相应的 Peer 状态。Normal 表示此 Peer 正常，Tombstone 表示此 Peer 已从 Region 中移除，不能再加入 Raft 组。

这些状态存储在两个 badger 实例中：raftdb 和 kvdb：

- raftdb 存储 raft 日志和 `RaftLocalState`
- kvdb 以不同的列族存储键值数据、`RegionLocalState` 和 `RaftApplyState`。你可以将 kvdb 视为 Raft 论文中提到的状态机。

格式如下，`kv/raftstore/meta` 中提供了一些辅助函数，并使用 `writebatch.SetMeta()` 将它们设置到 badger。

| Key              | KeyFormat                        | Value            | DB   |
| :--------------- | :------------------------------- | :--------------- | :--- |
| raft_log_key     | 0x01 0x02 region_id 0x01 log_idx | Entry            | raft |
| raft_state_key   | 0x01 0x02 region_id 0x02         | RaftLocalState   | raft |
| apply_state_key  | 0x01 0x02 region_id 0x03         | RaftApplyState   | kv   |
| region_state_key | 0x01 0x03 region_id 0x01         | RegionLocalState | kv   |

> 你可能想知道为什么 TinyKV 需要两个 badger 实例。实际上，它可以使用一个 badger 来存储 raft 日志和状态机数据。分成两个实例只是为了与 TiKV 的设计保持一致。

这些元数据应该在 `PeerStorage` 中创建和更新。创建 PeerStorage 时，请参见 `kv/raftstore/peer_storage.go`。它初始化此 Peer 的 RaftLocalState、RaftApplyState，或者在重启的情况下从底层引擎获取先前的值。请注意，RAFT_INIT_LOG_TERM 和 RAFT_INIT_LOG_INDEX 的值都是 5（只要大于 1 即可），而不是 0。不将其设置为 0 的原因是为了与在 conf change 后被动创建的 peer 的情况区分开来。你现在可能不太理解，只需记住这一点，细节将在 project3b 实现 conf change 时描述。

你在这部分需要实现的代码只有一个函数：`PeerStorage.SaveReadyState`，这个函数的作用是将 `raft.Ready` 中的数据保存到 badger，包括附加日志条目和保存 Raft 的硬状态。

要附加日志条目，只需将 `raft.Ready.Entries` 中的所有日志条目保存到 raftdb，并删除任何先前附加但永远不会被提交的日志条目。同时，更新 peer storage 的 `RaftLocalState` 并将其保存到 raftdb。

保存硬状态也非常简单，只需更新 peer storage 的 `RaftLocalState.HardState` 并将其保存到 raftdb。

> 提示：
>
> - 使用 `WriteBatch` 一次性保存这些状态。
> - 查看 `peer_storage.go` 中的其他函数，了解如何读写这些状态。
> - 设置环境变量 `LOG_LEVEL=debug` 可能有助于你调试，另请参阅所有可用的[日志级别](../log/log.go)。

### 实现 Raft ready 处理

在 project2 的 A 部分，你已经构建了一个基于 tick 的 Raft 模块。现在你需要编写外部流程来驱动它。大部分代码已经在 `kv/raftstore/peer_msg_handler.go` 和 `kv/raftstore/peer.go` 下实现。所以你需要学习这些代码并完成 `proposeRaftCommand` 和 `HandleRaftReady` 的逻辑。以下是对该框架的一些解释。

Raft `RawNode` 已经用 `PeerStorage` 创建并存储在 `peer` 中。在 raft worker 中，你可以看到它获取 `peer` 并用 `peerMsgHandler` 包装它。`peerMsgHandler` 主要有两个函数：一个是 `HandleMsg`，另一个是 `HandleRaftReady`。

`HandleMsg` 处理从 raftCh 收到的所有消息，包括调用 `RawNode.Tick()` 来驱动 Raft 的 `MsgTypeTick`，包装来自客户端请求的 `MsgTypeRaftCmd`，以及在 Raft peer 之间传输的消息 `MsgTypeRaftMessage`。所有的消息类型都在 `kv/raftstore/message/msg.go` 中定义。你可以查看它以了解详情，其中一些将在后续部分使用。

消息处理后，Raft 节点应该会有一些状态更新。所以 `HandleRaftReady` 应该从 Raft 模块获取 ready 并执行相应的操作，比如持久化日志条目、应用已提交的条目以及通过网络向其他 peer 发送 raft 消息。

用伪代码来说，raftstore 使用 Raft 的方式如下：

``` go
for {
  select {
  case <-s.Ticker:
    Node.Tick()
  default:
    if Node.HasReady() {
      rd := Node.Ready()
      saveToStorage(rd.State, rd.Entries, rd.Snapshot)
      send(rd.Messages)
      for _, entry := range rd.CommittedEntries {
        process(entry)
      }
      s.Node.Advance(rd)
    }
}
```

此后，一次读或写的完整过程将是这样的：

- 客户端调用 RPC RawGet/RawPut/RawDelete/RawScan
- RPC 处理程序调用 `RaftStorage` 相关方法
- `RaftStorage` 向 raftstore 发送一个 Raft 命令请求，并等待响应
- `RaftStore` 将 Raft 命令请求作为 Raft 日志提议
- Raft 模块附加日志，并通过 `PeerStorage` 持久化
- Raft 模块提交日志
- Raft worker 在处理 Raft ready 时执行 Raft 命令，并通过回调返回响应
- `RaftStorage` 从回调接收响应并返回给 RPC 处理程序
- RPC 处理程序执行一些操作并向客户端返回 RPC 响应。

你应该运行 `make project2b` 来通过所有测试。整个测试运行一个模拟集群，包括多个 TinyKV 实例和一个模拟网络。它执行一些读写操作并检查返回值是否符合预期。

需要注意的是，错误处理是通关测试的重要部分。你可能已经注意到 `proto/proto/errorpb.proto` 中定义了一些错误，并且错误是 gRPC 响应的一个字段。此外，`kv/raftstore/util/error.go` 中定义了实现 `error` 接口的相应错误，所以你可以将它们用作函数的返回值。

这些错误主要与 Region 相关。所以它也是 `RaftCmdResponse` 的 `RaftResponseHeader` 的一个成员。在提议请求或应用命令时，可能会出现一些错误。如果出现错误，你应该返回带有错误的 raft 命令响应，然后该错误将进一步传递给 gRPC 响应。当返回带有错误的响应时，你可以使用 `kv/raftstore/cmd_resp.go` 中提供的 `BindRespError` 将这些错误转换为 `errorpb.proto` 中定义的错误。

在这个阶段，你可能需要考虑这些错误，其他的将在 project3 中处理：

- ErrNotLeader: raft 命令在一个 follower 上被提议。所以用它来让客户端尝试其他 peer。
- ErrStaleCommand: 这可能是由于 leader 变更导致一些日志未被提交并被新 leader 的日志覆盖。但客户端不知道这一点，仍在等待响应。所以你应该返回这个错误，让客户端知道并重试该命令。

> 提示：
>
> - `PeerStorage` 实现了 Raft 模块的 `Storage` 接口，你应该使用提供的 `SaveReadyState()` 方法来持久化 Raft 相关的状态。
> - 使用 `engine_util` 中的 `WriteBatch` 来原子地进行多次写入，例如，你需要确保在一次写批处理中应用已提交的条目并更新已应用的索引。
> - 使用 `Transport` 向其他 peer 发送 raft 消息，它在 `GlobalContext` 中。
> - 如果服务器不属于多数派且没有最新的数据，则不应完成 get RPC。你可以将 get 操作放入 raft 日志，或者实现 Raft 论文第 8 节中描述的只读操作优化。
> - 应用日志条目时，不要忘记更新和持久化应用状态。
> - 你可以像 TiKV 那样异步地应用已提交的 Raft 日志条目。这不是必须的，但对于提高性能是一个很大的挑战。
> - 提议时记录命令的回调，并在应用后返回回调。
> - 对于 snap 命令响应，应将 badger Txn 显式设置到回调中。
> - 在 2A 之后，你可能需要多次运行一些测试来发现 bug。

## Part C

就目前你的代码而言，让一个长期运行的服务器永远记住完整的 Raft 日志是不切实际的。相反，服务器会检查 Raft 日志的数量，并时常丢弃超过阈值的日志条目。

在这一部分，你将在以上两部分实现的基础上实现快照处理。通常，快照就像 AppendEntries 一样，只是一个用于向 follower 复制数据的 raft 消息，不同之处在于它的大小。快照包含了某个时间点的整个状态机数据，一次性构建和发送如此大的消息会消耗大量资源和时间，这可能会阻塞其他 raft 消息的处理。为了分摊这个问题，快照消息将使用一个独立的连接，并将数据分成块来传输。这就是为什么 TinyKV 服务有一个快照 RPC API 的原因。如果你对发送和接收的细节感兴趣，请查看 `snapRunner` 和参考资料 <https://pingcap.com/blog-cn/tikv-source-code-reading-10/>

### 代码

你需要做的所有更改都基于在 A 部分和 B 部分编写的代码。

### 在 Raft 中实现

尽管我们需要对快照消息进行一些不同的处理，但从 raft 算法的角度来看，应该没有区别。查看 proto 文件中 `eraftpb.Snapshot` 的定义，`eraftpb.Snapshot` 上的 `data` 字段并不代表实际的状态机数据，而是一些供上层应用程序使用的元数据，你现在可以忽略它。当 leader 需要向 follower 发送快照消息时，它可以调用 `Storage.Snapshot()` 来获取一个 `eraftpb.Snapshot`，然后像其他 raft 消息一样发送快照消息。状态机数据实际上是如何构建和发送的由 raftstore 实现，这将在下一步介绍。你可以假设一旦 `Storage.Snapshot()` 成功返回，Raft leader 就可以安全地将快照消息发送给 follower，follower 应该���用 `handleSnapshot` 来处理它，也就是从消息中的 `eraftpb.SnapshotMetadata` 恢复 raft 的内部状态，如任期、提交索引和成员信息等，之后，快照处理过程就完成了。

### 在 raftstore 中实现

在这一步，你需要学习 raftstore 的另外两个 worker——raftlog-gc worker 和 region worker。

Raftstore 根据配置 `RaftLogGcCountLimit` 定期检查是否需要 gc 日志，请参见 `onRaftGcLogTick()`。如果是，它将提议一个 raft admin 命令 `CompactLogRequest`，该命令被包装在 `RaftCmdRequest` 中，就像在 project2 B 部分中实现的四种基本命令类型（Get/Put/Delete/Snap）一样。然后你需要在 Raft 提交此 admin 命令时处理它。但与 Get/Put/Delete/Snap 命令写入或读取状态机数据不同，`CompactLogRequest` 修改元数据，即更新 `RaftApplyState` 中的 `RaftTruncatedState`。之后，你应该通过 `ScheduleCompactLog` 向 raftlog-gc worker 调度一个任务。Raftlog-gc worker 将异步地执行实际的日志删除工作。

然后由于日志压缩，Raft 模块可能需要发送快照。`PeerStorage` 实现了 `Storage.Snapshot()`。TinyKV 在 region worker 中生成和应用快照。当调用 `Snapshot()` 时，它实际上向 region worker 发送一个任务 `RegionTaskGen`。region worker 的消息处理程序位于 `kv/raftstore/runner/region_task.go`。它扫描底层引擎以生成快照，并通过通道发送快照元数据。下次 Raft 调用 `Snapshot` 时，它会检查快照生成是否完成。如果是，Raft 应该将快照消息发送给其他 peer，快照的发送和接收工作由 `kv/storage/raft_storage/snap_runner.go` 处理。你不需要深入了解细节，只需要知道快照消息在快照接收后将由 `onRaftMsg` 处理。

然后快照将反映在下一次 Raft ready 中，所以你应该做的任务是修改 raft ready 过程以处理快照的情况。当你确定要应用快照时，你可以更新 peer storage 的内存状态，如 `RaftLocalState`、`RaftApplyState` 和 `RegionLocalState`。另外，不要忘记将这些状态持久化到 kvdb 和 raftdb，并从 kvdb 和 raftdb 中删除过时的状态。此外，你还需要将 `PeerStorage.snapState` 更新为 `snap.SnapState_Applying`，并通过 `PeerStorage.regionSched` 向 region worker 发送 `runner.RegionTaskApply` 任务，并等待 region worker 完成。

你应该运行 `make project2c` 来通过所有测试。
