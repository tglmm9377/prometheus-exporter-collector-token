# Prometheus-exporter-collector-token
作为Nightingale的插件，用于收集prometheus的指标

基于nightingale Prometheus-exporter-collector，根据实际使用场景做了一些调整

>1.支持 https  
>2支持k8s 带有token 认证的metrics接口  
>3.使用时候将bearer_token文件放到Prometheus-exporter-collector同级目录下
内容替换为你自己的token

```
func gatherExporterUrl(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		//Timeout: time.Duration(config.Get().Timeout) * time.Millisecond,
		Transport: tr,
	}
	var buf []byte
	//var req *http.Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil{
		fmt.Println("new request error",err)
		return nil, err
	}
	tokenbytes ,err := ioutil.ReadFile("bearer_token")
```
## 快速构建 

    $ mkdir -p $GOPATH/src/github.com/n9e
    $ cd $GOPATH/src/github.com/n9e
    $ git clone https://github.com/n9e/prometheus-exporter-collector.git
    $ cd prometheus-exporter-collector
    $ export GO111MODULE=on
    $ export GOPROXY=https://goproxy.cn
    $ go build
    $ cat plugin.test.json | ./prometheus-exporter-collector 


 ### 配置参数
 Name                             |  type     | Description
 ---------------------------------|-----------|--------------------------------------------------------------------------------------------------
 exporter_urls                    | array     | Address to collect metric for prometheus exporter.
 append_tags                      | array     | Append tags for n9e metric default empty
 endpoint                         | string    | Field endpoint for n9e metric default empty
 ignore_metrics_prefix            | array     | Ignore metric prefix default empty
 timeout                          | int       | Timeout for access a exporter url default 500ms
 metric_prefix                    | string    | append metric prefix when push to n9e. e.g. 'xx_exporter.'
 metric_type                      | map       | specify metric type
 default_mapping_metric_type      | string    | Default conversion rule for Prometheus cumulative metrics. support COUNTER and SUBTRACT. default SUBTRACT
 ###
 
 ###

```
$ cat plugin.test.json

{
  "exporter_urls": [
    "http://127.0.0.1:9121/metrics"
  ],
  "append_tags": ["region=bj", "dept=cloud"],
  "endpoint": "127.0.0.100",
  "ignore_metrics_prefix": ["go_"],
  "metric_prefix": "",
  "metric_type": {},
  "default_mapping_metric_type": "COUNTER",
  "timeout": 500
}
```

- 运行prometheus-exporter-collector，将输出发送给本机的 falcon-agent

```
cat plugin.test.json | ./prometheus-exporter-collector -b falcon -s 60 | curl -X POST -d @- http://127.0.0.1:1988/v1/push
```
