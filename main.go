package main

import (
	"context"
	"crypto/tls"
	"dlpu-rpc/frp"
	"dlpu-rpc/model"
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"

	"github.com/smallnest/rpcx/protocol"

	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
)

//server_addr 包含frp和etcd，分别是2379和7000
var (
	publicAddr = flag.String("public_addr", "", "服务器公网ip")
	serverAddr = flag.String("server_addr", "", "服务地址")
	token      = flag.String("token", "", "token")
	frpToken   = flag.String("frp_token", "", "frp的token")
	basePath   = "/rpc_jwc"
)

func init() {
	flag.Parse()
	fmt.Println(*serverAddr, *publicAddr, *token)
	var err error
	err = frp.Init(*publicAddr+":7000", *frpToken)
	if err != nil {
		logrus.Fatal("frp初始化错误", err)
	}
	logrus.Info("frp初始化成功")

}

func main() {
	var err error
	addr := "127.0.0.1:8972"

	cert, err := tls.X509KeyPair(pem, key)
	if err != nil {
		logrus.Panicln("证书解析失败", err)
	}
	//
	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	s := server.NewServer(server.WithTLSConfig(config))

	fmt.Println(`将要在` + *serverAddr + ":" + strconv.Itoa(frp.RandPort) + `运行服务`)
	addRegistryPlugin(s)
	err = s.RegisterName("JWC", new(model.User), "")
	if err != nil {
		panic(err)
	}
	s.AuthFunc = auth

	errs := s.Serve("tcp", addr)
	if errs != nil {
		fmt.Print(errs)
		panic(`服务出错，停止服务`)
	}
}

func addRegistryPlugin(s *server.Server) {

	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + *serverAddr + ":" + strconv.Itoa(frp.RandPort),
		EtcdServers:    []string{*publicAddr + ":2379"},
		BasePath:       basePath,
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}

func auth(ctx context.Context, req *protocol.Message, t string) error {

	if *token == t {
		return nil
	}

	return errors.New("invalid token")
}
