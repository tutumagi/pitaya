#### 1000 个对象，同时移动，进出

##### 20200629

```
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 69.246624ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 75.969046ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 86.671074ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 70.191131ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 70.641907ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 73.850643ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 71.005792ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 71.767157ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 70.187941ms
    TestXZListAOIManager/XZAOI: aoi_test.go:79: tick 1000 objects takes 70.826978ms
```


##### 20200629 去掉了 data2node 的map查询

```
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 33.175331ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 33.854003ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 32.360992ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 31.708987ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 32.592416ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 33.502855ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 34.78498ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 31.545924ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 35.235139ms
    TestXZListAOIManager: aoi_test.go:90: tick 1000 objects takes 36.563763ms
```