
分布式缓存小demo

启动三个数据存储节点
```bash
./server -addr=localhost:9090
./server -addr=localhost:9091
./server -addr=localhost:9092
```
然后在client set和get。观察server的数据情况