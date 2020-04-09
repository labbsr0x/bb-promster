package main

import (
	"text/template"
	"fmt"
	"os"
	"os/exec"
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

	
	_, err = cli.Put(context.TODO(), "/pilot_version/v0002", "")
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

	defer cli.Close()

	versionsChan := make(chan Version)
	
	go versions.watchUpdatedVersions(cli, versionsChan)
	
	for {
		select{
		case versions := <-versionsChan:
			fmt.Println(versions)
			generateAlertFile(versions)
			updatePrometheus()
		}
	}
	
}

func generateAlertFile(v Version) {
	fmt.Println("Generate alert file")
	// templating 
	tmpl, err := template.ParseFiles("/etc/prometheus/alert-rules.yml.tmpl")
	if err != nil { panic(err) }

	// file creating
	f, err := os.Create("/etc/prometheus/comparative-alerts.yml")
	if err != nil {
		fmt.Println("create file: ", err)
		return
	}
	err = tmpl.Execute(f, v)
	if err != nil { panic(err) }
}

func (v Version) watchUpdatedVersions(cli *clientv3.Client, versionsChan chan Version) {
	watchChan := cli.Watch(context.TODO(), "/prod_version", clientv3.WithPrefix())
	
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
			
			
			versionsChan <- v	
		}
		<- watchChan
		time.Sleep(time.Second)
	}
}

func updatePrometheus() {
	fmt.Println("Updating prometheus alert files")
	cmd := exec.Command("wget", "--post-data=''", "http://localhost:9090/-/reload", "-O", "-")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}