# Project1 独立式键值存储

在这个项目中，你将构建一个支持列族的独立键值存储 [gRPC](https://grpc.io/docs/guides/) 服务。独立式意味着只有单个节点，而不是分布式系统。[列族](https://en.wikipedia.org/wiki/Standard_column_family)（下文简称CF）是一个类似键命名空间的概念，即相同键在不同列族中的值是不同的。你可以简单地将多个列族视为独立的小型数据库。它用于支持项目4中的事务模型，到时你将了解为什么TinyKV需要CF的支持。

该服务支持四种基本操作：Put/Delete/Get/Scan。它维护一个简单的键值对数据库。键和值都是字符串。`Put`替换数据库中指定CF的特定键的值，`Delete`删除指定CF中键的值，`Get`获取指定CF中键的当前值，`Scan`获取指定CF中一系列键的当前值。

该项目可以分为2个步骤：

1. 实现一个独立的存储引擎。
2. 实现原始键值服务处理程序。

### 代码结构

`gRPC`服务器在`kv/main.go`中初始化，它包含一个提供名为`TinyKv`的`gRPC`服务的`tinykv.Server`。它是通过`proto/proto/tinykvpb.proto`中的[protocol-buffer](https://developers.google.com/protocol-buffers)定义的，rpc请求和响应的详细信息在`proto/proto/kvrpcpb.proto`中定义。

通常，你不需要更改proto文件，因为所有必要的字段都已为你定义好了。但如果你仍然需要更改，可以修改proto文件并运行`make proto`来更新`proto/pkg/xxx/xxx.pb.go`中相关的生成的go代码。

此外，`Server`依赖于一个`Storage`，这是你需要在`kv/storage/standalone_storage/standalone_storage.go`中为独立存储引擎实现的接口。一旦在`StandaloneStorage`中实现了`Storage`接口，你就可以用它为`Server`实现原始键值服务。

#### 实现独立存储引擎

第一个任务是实现[badger](https://github.com/dgraph-io/badger)键值API的包装器。gRPC服务器的服务依赖于在`kv/storage/storage.go`中定义的`Storage`。在这种情况下，独立存储引擎只是badger键值API的包装器，它由两个方法提供：

```go
type Storage interface {
    // 其他内容
    Write(ctx *kvrpcpb.Context, batch []Modify) error
    Reader(ctx *kvrpcpb.Context) (StorageReader, error)
}
```

`Write`应该提供一种方法，将一系列修改应用到内部状态，在这种情况下是badger实例。

`Reader`应该返回一个`StorageReader`，它支持在快照上进行键值的点查询和扫描操作。

现在你不需要考虑`kvrpcpb.Context`，它将在后续项目中使用。

> 提示：
>
> - 你应该使用[badger.Txn](https://godoc.org/github.com/dgraph-io/badger#Txn)来实现`Reader`函数，因为badger提供的事务处理程序可以提供键和值的一致性快照。
> - Badger不支持列族。engine_util包（`kv/util/engine_util`）通过向键添加前缀来模拟列族。例如，属于特定列族`cf`的键`key`存储为`${cf}_${key}`。它包装了`badger`以提供带有CF的操作，并且还提供了许多有用的辅助函数。因此，你应该通过`engine_util`提供的方法进行所有读/写操作。请阅读`util/engine_util/doc.go`了解更多信息。
> - TinyKV使用了原始版本`badger`的一个分支，其中包含一些修复，所以请使用`github.com/Connor1996/badger`而不是`github.com/dgraph-io/badger`。
> - 不要忘记在丢弃之前为badger.Txn调用`Discard()`并关闭所有迭代器。

#### 实现服务处理程序

本项目的最后一步是使用已实现的存储引擎构建原始键值服务处理程序，包括RawGet/RawScan/RawPut/RawDelete。处理程序已经为你定义好了，你只需要在`kv/server/raw_api.go`中填写实现即可。完成后，记得运行`make project1`以通过测试套件。
