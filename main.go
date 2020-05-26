package main

import (
	"context"
	"os"
	"os/exec"
	"path"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
)

type Label struct {
	PilotVersion string
	ProdVersion  string
	Prsn         string
}

func main() {
	viper.AutomaticEnv()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{viper.GetString("ETCD_URLS")},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logrus.Fatal("Could not connect to etcd")
		panic(err)
	}

	defer cli.Close()

	var labels Label
	labelsChan := make(chan Label)

	go labels.watchUpdatedVersions(cli, labelsChan)

	for {
		select {
		case versions := <-labelsChan:
			generateAlertFile(versions)
			updatePrometheus()
		}
	}

}

func generateAlertFile(v Label) {
	logrus.Info("Generate alert file")

	tmpl, err := template.ParseFiles("/etc/prometheus/alert-rules.yml.tmpl")
	if err != nil {
		panic(err)
	}

	f, err := os.Create("/etc/prometheus/comparative-alerts.yml")
	if err != nil {
		panic(err)

	}
	err = tmpl.Execute(f, v)
	if err != nil {
		panic(err)
	}
}

func (v Label) watchUpdatedVersions(cli *clientv3.Client, versionsChan chan Label) {
	watchChan := cli.Watch(context.TODO(), "/versions", clientv3.WithPrefix())

	for {
		rspProd, err := cli.Get(context.TODO(), "/versions/"+viper.GetString("REGISTRY_SERVICE")+"/prod_version", clientv3.WithPrefix())
		rspPilot, err := cli.Get(context.TODO(), "/versions/"+viper.GetString("REGISTRY_SERVICE")+"/pilot_version", clientv3.WithPrefix())
		if err != nil {
			panic(err)
		}

		if len(rspProd.Kvs) == 0 || len(rspPilot.Kvs) == 0 {
			logrus.Warn("Pilot or Prod version not found")
		} else {
			prodPath := string(rspProd.Kvs[len(rspProd.Kvs)-1].Key)
			pilotPath := string(rspPilot.Kvs[len(rspPilot.Kvs)-1].Key)
			v.ProdVersion = path.Base(prodPath)
			v.PilotVersion = path.Base(pilotPath)
			v.Prsn = viper.GetString("REGISTRY_SERVICE")
			versionsChan <- v
		}
		<-watchChan
	}
}

func updatePrometheus() {
	logrus.Info("Updating prometheus alert files")
	cmd := exec.Command("wget", "--post-data=''", "http://localhost:9090/-/reload", "-O", "-")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
