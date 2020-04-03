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
	
	//starting version struct
	versions := Version {
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

	versionsChan := make(chan Version)
	generateAlertFile(pilot_version, prod_version)
	
	go versions.watchUpdatedVersions(cli, versionsChan)
	
	for {
		select{
		case versions := <-versionsChan:
			fmt.Println(versions)
		}
	}
	
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

func (v Version) watchUpdatedVersions(cli *clientv3.Client, versionsChan chan Version) {
	//fmt.Println("Version chan", <-versionsChan)
	for {
		rspProd, err0 := cli.Get(context.TODO(), "/prod_version", clientv3.WithPrefix())
		rspPilot, err0 := cli.Get(context.TODO(), "/pilot_version", clientv3.WithPrefix())
		
		if err0 != nil {
			fmt.Print(err0)
		}
		
		if len(rspProd.Kvs) == 0 || len(rspPilot.Kvs) == 0{
			fmt.Println("Pilot or Prod version not found")
		} else {
			prodPath := string(rspProd.Kvs[len(rspProd.Kvs)-1].Key)
			pilotPath := string(rspPilot.Kvs[len(rspPilot.Kvs)-1].Key)
			v.ProdVersion = path.Base(prodPath)
			v.PilotVersion = path.Base(pilotPath)
			fmt.Println("Prod version", v.ProdVersion)
			fmt.Println("Pilot version", v.PilotVersion)
			versionsChan <- v
			
		}
		
		//fmt.Println("Version chan", <-versionsChan)
		time.Sleep(5 * time.Second)
	}
}