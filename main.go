package main

import (
	"text/template"
	"fmt"
	"os"
	"go.etcd.io/etcd/clientv3"
	"time"
	"context"
)

type Version struct{
	PilotVersion string
	ProdVersion string
}

func main() {
	// templating 
	versions := Version{"v0002", "v0001"}
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

	value, err := cli.Get(context.TODO(), "/pilot_version", clientv3.WithPrefix())
	fmt.Print(value)

	if err != nil {
    	fmt.Print(err)
	}

	defer cli.Close()
}