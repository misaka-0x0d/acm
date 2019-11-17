package acm

import (
	"errors"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/bilibili/kratos/pkg/conf/paladin"
	"sync"

	"context"
	"github.com/bilibili/kratos/pkg/log"
	"os"
	"strings"
)
//https://github.com/nacos-group/nacos-sdk-go
var (
	zoneId 		string
	endpoint 	string
    namespaceId string

	accessKey 	string
	secretKey 	string
	group 		string
	dataId 		string
)

type Config struct {
	ZoneId 		string
	Endpoint 	string
	NamespaceId string
	AccessKey   string
	SecretKey   string
	Group		string
	DataId		[]string
}

type aliAcm struct {
	clinet *config_client.IConfigClient
	values *paladin.Map
	wmu sync.RWMutex
	watchers map[*ACMWatcher] struct{}
}

type ACMWatcher struct {
	keys []string
	C chan paladin.Event
}

func (aw *ACMWatcher)HasKey(key string) bool{
	if len(aw.keys) == 0{
		return true
	}
	for _, k := range aw.keys{
		if k == key{
			return true
		}
	}
	return false
}

func (aw *ACMWatcher)Handle(event paladin.Event)  {
	select {
	case aw.C <- event:
	default:
		log.Warn("paladin: event channel full discard ns %s update event", event.Key)
	}
}

func newACMWatcher(keys []string) *ACMWatcher{
	return &ACMWatcher{
		keys: keys,
		C:    make(chan paladin.Event, 5),
	}
}

type acmDriver struct{}

func addFlags()  {
	flag.StringVar(&zoneId, "acm.zone", "", "Zone ID where ACM is located")
	flag.StringVar(&endpoint, "acm.endpoint", "", "ACM's endpoint address")
	flag.StringVar(&namespaceId, "acm.namespace", "", "ACM's namespaceId")
	flag.StringVar(&accessKey, "acm.accesskey", "", "RAM AccessKey")
	flag.StringVar(&secretKey, "acm.secretKey", "", "RAM secretKey")
	flag.StringVar(&group, "acm.group", "", "Group name of the ACM profile")
	flag.StringVar(&dataId, "acm.data", "", "ACM configuration set ID,Use \",\" to separate when there are multiple configurations")
}

func init() {
	addFlags()
	paladin.Register(PaladinDriverAliyunACM,  &acmDriver{})
}

func buildAcmConfig() (conf *Config, err error){
	if acmZoneId := os.Getenv("ACM_ZONE_ID"); acmZoneId != ""{
		zoneId = acmZoneId
	}
	//if zoneId == "" {
	//	err = errors.New("invalid ACM zoneId, pass it via ACM_ZONE_ID=xxx with env or --acm.zone=xxx with flag")
	//	return
	//}

	if acmEndpoint := os.Getenv("ACM_ENDPOINT_ADDR"); acmEndpoint != ""{
		endpoint = acmEndpoint
	}
	if endpoint == "" {
		err = errors.New("invalid ACM endpoint, pass it via ACM_ENDPOINT_ADDR=xxx with env or --acm.endpoint=xxx with flag")
		return
	}

	if acmNamespaceId := os.Getenv("ACM_NAMESPACE_ID"); acmNamespaceId != ""{
		namespaceId = acmNamespaceId
	}
	if namespaceId == ""{
		err = errors.New("invalid ACM namespaceId, pass it via ACM_NAMESPACE_ID=xxx with env or --acm.namespace=xxx with flag")
		return
	}

	if acmAccessKey := os.Getenv("ACCESS_KEY"); acmAccessKey != ""{
		accessKey = acmAccessKey
	}
	if accessKey == ""{
		err = errors.New("invalid accessKey, pass it via ACCESS_KEY=xxx with env or --acm.accesskey=xxx with flag")
		return
	}

	if acmSecretKey := os.Getenv("SECRET_KEY"); acmSecretKey != ""{
		secretKey = acmSecretKey
	}
	if secretKey == ""{
		err = errors.New("invalid secretKey, pass it via SECRET_KEY=xxx with env or --acm.secretKey=xxx with flag")
		return
	}

	if acmGroup := os.Getenv("ACM_GROUP"); acmGroup != ""{
		group = acmGroup
	}
	if group == ""{
		err = errors.New("invalid ACM group, pass it via ACM_GROUP=xxx with env or --acm.group=xxx with flag")
		return
	}

	if acmData := os.Getenv("ACM_DATA"); acmData != ""{
		dataId = acmData
	}
	dataIds := strings.Split(dataId, ",")
	if len(dataIds) == 0{
		err = errors.New("invalid ACM dataId, pass it via ACM_DATA=xxx with env or --acm.data=xxx with flag")
		return
	}

	conf = &Config{
		ZoneId:      zoneId,
		Endpoint:    endpoint,
		NamespaceId: namespaceId,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Group:       group,
		DataId:      dataIds,
	}
	return
}

func (acm *acmDriver)New() (paladin.Client, error){
	conf, err := buildAcmConfig()
	if err != nil {
		return nil, err
	}
	return acm.new(conf)
}

func (acm *acmDriver)new(conf *Config) (paladin.Client, error){
	if conf == nil{
		return nil, errors.New("invalid Aliyun ACM conf")
	}
	clientConfig := constant.ClientConfig{
		Endpoint:       endpoint + ":8080",
		NamespaceId:    namespaceId,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		TimeoutMs:      5 * 1000,
		ListenInterval: 30 * 1000,
	}
	//建立ACM客户端
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"clientConfig": clientConfig,
	})
	if err != nil{
		return nil, err
	}
	a := &aliAcm{
		clinet: &configClient,
		values: new(paladin.Map),
		watchers: make(map[*ACMWatcher]struct{}),
	}
	configStr, err := a.loadConfig(conf.Group, conf.DataId)
	if err != nil{
		return nil, err
	}
	a.values.Store(configStr)
	a.WatchEvent(context.TODO(), conf.DataId...)
	//监听全部配置
	for _, k := range conf.DataId{
		err := configClient.ListenConfig(vo.ConfigParam{
			DataId: k,
			Group:  conf.Group,
			OnChange: a.listenCallback,
		})
		if err != nil{
			log.Error("ListenConfig G=%s,D=%s failed(err=%s)",conf.Group, k, err.Error())
			return nil, err
		}
	}
	return a, nil
}

// 加载配置
func (a *aliAcm) loadConfig(g string, d []string) (values map[string]*paladin.Value, err error){
	values = make(map[string]*paladin.Value, len(d))
	for _, k := range d{
		var str string
		clinet := *a.clinet
		str, err = clinet.GetConfig(vo.ConfigParam{DataId: k, Group:  g});
		if err != nil{
			log.Error("paladin: ACM read config %s failed(err=%s)", k, err.Error())
			return
		}
		values[k], err = a.conversionData(str);
		if err != nil{
			return
		}
	}
	return
}

// 转换paladin.Value
func (a *aliAcm) conversionData(data string) (*paladin.Value, error){
	return paladin.NewValue(data, data), nil
}

// 配置监听回调
func (a *aliAcm)listenCallback(namespace, group, dataId, data string){
	log.Warn("paladin: listen that configuration G=%s,D=%s has changed", group, dataId)
	value := &paladin.Value{}
	value = paladin.NewValue(data, data)
	raws := a.values.Load()
	raws[dataId] = value
	rawValue, err := value.Raw()
	if err != nil {
		return
	}
	a.values.Store(raws)
	a.wmu.RLock()
	n := 0
	for w := range a.watchers{
		if w.HasKey(dataId){
			n ++
			w.Handle(paladin.Event{
				Event: paladin.EventUpdate,
				Key: dataId,
				Value: rawValue,
			})
		}
	}
	a.wmu.RUnlock()
	log.Warn("paladin: reload config: %s events: %d\n", dataId, n)
	return
}

// 返回value的map
func (a *aliAcm) GetAll() *paladin.Map{
	return a.values
}

// 返回key的value
func (a *aliAcm) Get(key string) *paladin.Value{
	return a.values.Get(key)
}

//指定监听 keys
func (a *aliAcm) WatchEvent(ctx context.Context, keys ...string) <-chan paladin.Event {
	aw := newACMWatcher(keys)
	a.wmu.Lock()
	a.watchers[aw] = struct{}{}
	a.wmu.Unlock()
	return aw.C
}

//关闭配置监听
func (a *aliAcm) Close() (err error){
	a.wmu.RLock()
	for w := range a.watchers{
		close(w.C)
	}
	a.wmu.RUnlock()
	return nil
}
