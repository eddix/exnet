# exnet
exnet 通过扩展网络库的能力，提供在高并发服务场景下的网络连接解决方案。

## 主要功能：

* [x] 生成连接是一个完整实现的 `net.Conn`，可适用于各种标准库
* [x] 可以在Conn的各个环节添加callback
* [x] 从一组地址中根据负载均衡策略建立连接

## 使用示例

更多例子请看example目录。

### 直接使用

```go
// 添加名为echo的服务，以及它的IP列表
ap := addresspicker.NewRoundRobin(nil)
for _, srv := range []string{"192.168.0.1:8756", "192.168.0.2:8756"} {
    if err := ap.AppendTCPAddress("tcp", srv); err != nil {
        log.Fatal(err)
    }
}
cluster := exnet.Cluster{
    DialTimeout:   100 * time.Millisecond,
    ReadTimeout:   100 * time.Millisecond,
    WriteTimeout:  100 * time.Millisecond,
    AddressPicker: ap,
}

// 从exnet中获取echo服务的一个连接，network和address其实是不需要的，因为已经
// 加到 addresspicker。
conn, err := cluster.Dial("", "")
if err != nil {
    log.Fatalf("exnet:: Get Connection failed: %s\n")
}

// 查看这个连接的信息
log.Printf("exnet:: LocalAddr: %s, RemoteAddr: %s\n",
    conn.LocalAddr().String(), conn.RemoteAddr().String())

// 写入数据
conn.SetWriteDeadline(<-time.After(500 * time.Millisecond))
if _, err = conn.Write([]byte("hello")); err != nil {
    log.Fatalf("exnet:: Write failed")
}

// 读取数据
conn.SetReadDeadline(<-time.After(500*time.Millisecond))
buf := make([]byte, 5)
if _, err = conn.Read(buf); err != nil {
    log.Fatalf("exnet:: Read failed")
}

// 关闭连接
conn.Close()
```

### 在HTTP请求中使用

参考 example/http 示例。

### 在MySQL库中使用

参考 example/mysql 示例。

### 在Redis库中使用

参考 example/redis 示例。