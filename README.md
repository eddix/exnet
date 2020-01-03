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


### 在MySQL库中使用

```go
mysql.RegisterDialer("exnet", func(addr string) (net.Conn, error) {
    return cluster.Dial("", "")
})

db, err := sql.Open("mysql", "user:password@exnet(mydb)/dbname")
```

### 在Redis库中使用

#### 使用redigo/redis库

```go
import "github.com/gomodule/redigo/redis"

conn, _ := cluster.Dial("", "")
redisConn := redis.NewConn(conn, time.Second, time.Second)
_, err := redisConn.Do("PING")
```

#### 使用go-redis/redis库

```go
import "github.com/go-redis/redis"

client := redis.NewClient(&redis.Options{
    Addr: "myredis",
    DB: 0,
    Dialer: cluster.DialContext,
})

err := client.Set("key", "value", 0).Err()
if err != nil {
    log.Printf("client.Set error: %s\n", err.Error())
}
```
