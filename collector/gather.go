package collector

import (
	"crypto/tls"
	"fmt"
	"github.com/didi/nightingale/src/dataobj"
	"github.com/n9e/prometheus-exporter-collector/config"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

func Gather() []*dataobj.MetricValue {
	var wg sync.WaitGroup
	var res []*dataobj.MetricValue

	cfg := config.Get()
	metricChan := make(chan *dataobj.MetricValue)
	done := make(chan struct{}, 1)

	go func() {
		defer func() { done <- struct{}{} }()
		for m := range metricChan {
			res = append(res, m)
		}
	}()

	for _, exporterUrl := range cfg.ExporterUrls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if metrics, err := gatherExporter(url); err == nil {
				for _, m := range metrics {
					if typ, exists := cfg.MetricType[m.Metric]; exists {
						m.CounterType = typ
					}

					if cfg.MetricPrefix != "" {
						m.Metric = cfg.MetricPrefix + m.Metric
					}
					metricChan <- m
				}
			}
		}(exporterUrl)
	}

	wg.Wait()
	close(metricChan)

	<-done

	return res
}

func gatherExporter(url string) ([]*dataobj.MetricValue, error) {
	body, err := gatherExporterUrl(url)
	if err != nil {
		log.Printf("1. gather metrics from exporter error, url :[%s] ,error :%v", url, err)
		return nil, err
	}

	metrics, err := Parse(body)
	if err != nil {
		log.Printf("parse metrics error, url :[%s] ,error :%v", url, err)
		return nil, err
	}

	return metrics, nil
}

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
	if err != nil{
		fmt.Println("read token file error ",err)
		return nil ,err
	}
	req.Header.Set("Authorization","Bearer "+string(tokenbytes))
	if err != nil {
		return buf, err
	}


	//var resp *http.Response
	resp, err := client.Do(req)
	if err != nil {
		return buf, fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return buf, fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return buf, fmt.Errorf("error reading body: %s", err)
	}

	return body, nil
}
