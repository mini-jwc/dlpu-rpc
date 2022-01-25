package frp

import (
	"context"
	"fmt"
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/auth"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/consts"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/golib/crypto"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	RandPort = 10000 + rand.Intn(55534)
	frpAddr  string
	frpToken string
)

func Init(addr string, token string) (err error) {
	crypto.DefaultSalt = "frp"
	rand.Seed(time.Now().UnixNano())

	frpAddr = addr
	frpToken = token

	clientCfg, err := newClientCommonCfg()
	if err != nil {
		return
	}
	clientCfg.Complete()
	err = clientCfg.Validate()
	if err != nil {
		return
	}

	cfg := &config.TCPProxyConf{}
	var prefix string

	cfg.ProxyName = prefix + ""
	cfg.ProxyType = consts.TCPProxy
	//cfg.LocalIP = localIP
	cfg.LocalIP = "127.0.0.1"
	//cfg.LocalPort = localPort
	cfg.LocalPort = 8972
	cfg.RemotePort = RandPort
	cfg.UseEncryption = false
	cfg.UseCompression = false

	err = cfg.CheckForCli()
	if err != nil {
		return
	}

	proxyConfs := map[string]config.ProxyConf{
		cfg.ProxyName: cfg,
	}

	//fmt.Println("clientCfg", clientCfg)
	//fmt.Println("TCPProxyConf", cfg)

	err = startService(clientCfg, proxyConfs, nil, "")
	if err != nil {
		return
	}
	return nil
}

func startService(
	cfg config.ClientCommonConf,
	pxyCfgs map[string]config.ProxyConf,
	visitorCfgs map[string]config.VisitorConf,
	cfgFile string,
) error {

	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel,
		cfg.LogMaxDays, cfg.DisableLogColor)

	if cfg.DNSServer != "" {
		s := cfg.DNSServer
		if !strings.Contains(s, ":") {
			s += ":53"
		}
		// Change default dns server for frpc
		net.DefaultResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial("udp", s)
			},
		}
	}
	svr, errRet := client.NewService(cfg, pxyCfgs, visitorCfgs, cfgFile)
	if errRet != nil {
		return errRet
	}
	go func() {
		err := svr.Run()
		if err != nil {
			logrus.Fatal("frp运行错误")
		}
	}()

	return nil

}

func newClientCommonCfg() (cfg config.ClientCommonConf, err error) {
	cfg = config.GetDefaultClientConf()

	ipStr, portStr, err := net.SplitHostPort(frpAddr)
	if err != nil {
		err = fmt.Errorf("invalid server_addr: %v", err)
		return
	}

	cfg.ServerAddr = ipStr
	cfg.ServerPort, err = strconv.Atoi(portStr)
	if err != nil {
		err = fmt.Errorf("invalid server_addr: %v", err)
		return
	}

	cfg.User = ""
	cfg.Protocol = "tcp"
	cfg.LogLevel = "info"
	cfg.LogFile = "console"
	cfg.LogMaxDays = int64(3)

	// Only token authentication is supported in cmd mode
	cfg.ClientConfig = auth.GetDefaultClientConf()
	cfg.Token = frpToken
	cfg.TLSEnable = false

	return
}
