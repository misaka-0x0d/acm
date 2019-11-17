#### acm

##### 项目简介
`acm` 是blibili开源框架`Kratos` paladin组件的一个扩展

`
应用配置管理（Application Configuration Management，简称 ACM），其前身为淘宝内部配置中心 Diamond，是一款应用配置中心产品。基于该应用配置中心产品，您可以在微服务、DevOps、大数据等场景下极大地减轻配置管理的工作量的同时，保证配置的安全合规。
`

大部分代码参考自原有的Apollo扩展

| flag | ENV | 说明 |
|:------|:------|:------|
| acm.zone | ACM_ZONE_ID | ACM地域名称[可选],预留参数 |
| acm.endpoint | ACM_ENDPOINT_ADDR | endpoint节点地址,如:acm.aliyun.com |
| acm.namespace | ACM_NAMESPACE_ID | 配置所在地域节点的命名空间ID |
| acm.accesskey | ACCESS_KEY | 阿里云RAM用户 |
| acm.secretKey | SECRET_KEY | 阿里云RAM用户 |
| acm.group | ACM_GROUP | 配置组[Group] |
| acm.data | ACM_DATA | 配置集[Data ID], 多个配置使用","分割,如:http,grpc,db,memcache,redis,app |

##### 使用方法
```go
	if err := paladin.Init(acm.PaladinDriverAliyunACM); err != nil {
		panic(err)
	}
```
##### 编译环境

- **请只用 Golang v1.12.x 以上版本编译执行**

##### 依赖包
``` go
"github.com/nacos-group/nacos-sdk-go/clients"
"github.com/nacos-group/nacos-sdk-go/clients/config_client"
"github.com/nacos-group/nacos-sdk-go/common/constant"
"github.com/nacos-group/nacos-sdk-go/vo"
"github.com/bilibili/kratos/pkg/log"
```

