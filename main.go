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
	fmt.Print("Container up\n")
	versions := Version{"v0002", "v0001"}
	tmpl, err := template.ParseFiles("/etc/prometheus/alert-rules.yml.tmpl")
	if err != nil { panic(err) }
	err = tmpl.Execute(os.Stdout, versions)
	if err != nil { panic(err) }

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://etcd:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil { panic(err)}
	_, err = cli.Put(context.TODO(), "foo", "bar")
	if err != nil {
    	fmt.Print(err)
	}

	defer cli.Close()

	for {

	}
}