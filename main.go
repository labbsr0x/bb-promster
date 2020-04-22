package main

import (
	"text/template"
	"os"
	"os/exec"
	"time"
	"go.etcd.io/etcd/clientv3"
	"context"
	"path"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use: "bb-promster",
	Short: "A promster image definition to properly work with the Big Brother project",
}

type Version struct{
	PilotVersion string
	ProdVersion string
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
func init() {
	cobra.OnInitialize(initConfig)
	fmt.Println("Func called")
	viper.AutomaticEnv()
	fmt.Println("Testing around", viper.GetString("ETCD_URLS"))
}

func initConfig(){
	fmt.Println("FUNC CALLED")
	viper.SetEnvPrefix("REGISTRY_SERVICE")
	viper.AutomaticEnv()
	viper.AddRemoteProvider("etcd", "http://etcd:2379", "")
	viper.SetConfigType("json") // because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"
	err := viper.ReadRemoteConfig()
	if err != nil {
		logrus.Error(err)
	}
	
}

func main() {
	execute()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://etcd:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil { 
		logrus.Fatal("Could not connect to etcd")
		panic(err)
	}
	
	defer cli.Close()

	var versions Version 
	versionsChan := make(chan Version)
	
	go versions.watchUpdatedVersions(cli, versionsChan)
	
	for {
		select{
		case versions := <-versionsChan:
			generateAlertFile(versions)
			updatePrometheus()
		}
	}
	
}

func generateAlertFile(v Version) {
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

func (v Version) watchUpdatedVersions(cli *clientv3.Client, versionsChan chan Version) {
	watchChan := cli.Watch(context.TODO(), "/versions", clientv3.WithPrefix())
	
	for {
		rspProd, err := cli.Get(context.TODO(), "/versions/"+ os.Getenv("REGISTRY_SERVICE") + "/prod_version", clientv3.WithPrefix())
		rspPilot, err := cli.Get(context.TODO(), "/versions/"+ os.Getenv("REGISTRY_SERVICE") + "/pilot_version", clientv3.WithPrefix())
		if err != nil {
			panic(err) 
		}
		
		if len(rspProd.Kvs) == 0 || len(rspPilot.Kvs) == 0{
			logrus.Warn("Pilot or Prod version not found")
		} else {
			prodPath := string(rspProd.Kvs[len(rspProd.Kvs)-1].Key)
			pilotPath := string(rspPilot.Kvs[len(rspPilot.Kvs)-1].Key)
			v.ProdVersion = path.Base(prodPath)
			v.PilotVersion = path.Base(pilotPath)	
			versionsChan <- v	
		}
		<- watchChan
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