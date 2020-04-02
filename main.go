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
	
	go watchUpdatedVersions("/prod_version", cli)
	
	select{}
	
	
}

func generateAlertFile(pilot_version string, prod_version string) {
	fmt.Println("Generate file")
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

func forever() {
    for {
        fmt.Printf("%v+\n", time.Now())
        time.Sleep(time.Second)
    }
}

//it works, but isn't being called by go thread
func watchUpdatedVersionsDeprecated(version string, cli *clientv3.Client, versionsChan chan string){
	//for {
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
		//time.Sleep(time.Second)
	//} 
}

func watchUpdatedVersions(version string, cli *clientv3.Client) {
	for {
		rsp, err0 := cli.Get(context.TODO(), version, clientv3.WithPrefix())
		
		if err0 != nil {
			fmt.Print(err0)
		}
		
		if len(rsp.Kvs) == 0 {
			fmt.Printf("no %s versions founded\n", version)
		} else {	
			arraySize := len(rsp.Kvs)
			prodPath := string(rsp.Kvs[arraySize-1].Key)
			prodVersion := path.Base(prodPath)
			fmt.Println("Prod version", prodVersion)
			
			
			//versionsChan <- prod_version
		}
		
		time.Sleep(60 * time.Second)
	}
}