package main

import (
	"consual/etcdclient"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
)

const (
	Addr  = "192.168.200.101"
	Iport = 8089
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/product/{id:\\d+}", func(writer http.ResponseWriter, request *http.Request) {
		posts := mux.Vars(request)
		str := "get product ById" + posts["id"]
		writer.Write([]byte(str))
	})

	errChan := make(chan error)
	serviceId, serviceName, serviceAddr := "v1", "product", "192.168.200.101:8089"
	etcdClient, _ := etcdclient.NewEtcdClient()

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(Iport),
		Handler: router,
	}
	go (func() {
		err := etcdClient.RegisterService(serviceId, serviceName, serviceAddr)

		if err != nil {
			errChan <- err
			return
		}
		err = httpServer.ListenAndServe()

		if err != nil {
			errChan <- err
			return
		}
	})()

	go (func() {
		sigInterrupt := make(chan os.Signal)
		signal.Notify(sigInterrupt, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-sigInterrupt)
	})()

	getErr := <-errChan

	err := etcdClient.UnRegisterService(serviceId, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	err = httpServer.Shutdown(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("系统错误或被中断。。。。\n")
	log.Fatal(getErr)
}
