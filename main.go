package main

import (
	"text/template"
	"fmt"
	"os"
	"time"
	"go.etcd.io/etcd/clientv3"
	"context"
	"path"
)

type Version struct{
	PilotVersion string
	ProdVersion string
}

func main() {
	
	//etcd writing
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://etcd:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil { panic(err)}

	
	_, err = cli.Put(context.TODO(), "/pilot_version/v0003", "")
	if err != nil {
    	fmt.Print(err)
	}

	_, err = cli.Put(context.TODO(), "/prod_version/v0001", "")
	if err != nil {
    	fmt.Print(err)
	}
	
	
	//getting data from etcd
	kv, err := cli.Get(context.TODO(), "/pilot_version", clientv3.WithPrefix())
	pilot_version := path.Base(string(kv.Kvs[0].Key))

	kv2, err := cli.Get(context.TODO(), "/prod_version", clientv3.WithPrefix())
	prod_version := path.Base(string(kv2.Kvs[0].Key))

	if err != nil {
    	fmt.Print(err)
	}

	defer cli.Close()

	
	
	generateAlertFile(pilot_version, prod_version)
	versionsChan := make(chan string, 0)
	go watchUpdatedVersions("/prod_version", cli, versionsChan)
	//go watchVersions("/prod_version", cli, versionsChan)
	ver := <- versionsChan
	fmt.Print(ver)

	
}

func generateAlertFile(pilot_version string, prod_version string) {
	fmt.Print("Generate file")
	// templating 
	versions := Version{pilot_version, prod_version}
	tmpl, err := template.ParseFiles("/etc/prometheus/alert-rules.yml.tmpl")
	if err != nil { panic(err) }

	// file creating
	f, err := os.Create("/etc/prometheus/comparative-alerts.yml")
	if err != nil {
		fmt.Println("create file: ", err)
		return
	}
	err = tmpl.Execute(f, versions)
	if err != nil { panic(err) }
}

//it works, but isn't being called by go thread
func watchUpdatedVersions(version string, cli *clientv3.Client, versionsChan chan string){
	fmt.Print("function called")
	rch := cli.Watch(context.Background(), version)
	var prod_version string
	for wresp := range rch {
		fmt.Print(wresp)
		for _, ev := range wresp.Events {
			prod_version = string(ev.Kv.Key)
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
		versionsChan <- prod_version
	}	
}

func watchVersions(version string, cli *clientv3.Client, versionsChan chan string) {
	watchChan := cli.Watch(context.TODO(), version, clientv3.WithPrefix())
	for {
		fmt.Print("Versions updated")
		rsp, err0 := cli.Get(context.TODO(), version, clientv3.WithPrefix())
		if err0 != nil {
			fmt.Print(err0)
		}

		if len(rsp.Kvs) == 0 {
			fmt.Print("no prod versions founded")
		} else {
			var prod_version string
			for _, kv := range rsp.Kvs {
				fmt.Print("Versao atualizada")
				prod_version = path.Base(string(kv.Key))
			}
			versionsChan <- prod_version
		}
		<-watchChan
		time.Sleep(5 * time.Second)
	}
}