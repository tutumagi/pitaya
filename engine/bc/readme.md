base & cell 关于实体，场景切换，实体数据管理相关

base 表示 该实体当前创建在 跟 空间无关（不会触发aoi）的server
cell 表示 该实体当前创建在 跟 空间有关（会触发aoi）的server


### TODO
#### 实体属性，区分以下几种情况
1. 存在业务服的字段 base
2. 存在 aoi服的字段 cell
3. 是否需要通过 aoi 通知给周围玩家的字段 cellclient