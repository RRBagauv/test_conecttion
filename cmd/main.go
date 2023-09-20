package main

import "C"
import (
	net2 "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"os/signal"
	"runtime"
)

import (
	"bytes"
	"context"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
)

func init() {
	signals := make(chan os.Signal)
	signal.Notify(signals, unix.SIGUSR2)
	go func() {
		buf := make([]byte, os.Getpagesize())
		for {
			select {
			case <-signals:
				n := runtime.Stack(buf, true)
				buf[n] = 0
			}
		}
	}()
}

//export connect
func connect() {
	configJson := `{
  "dns": {
    "disableFallback": true,
    "servers": [
      {
        "address": "https://8.8.8.8/dns-query",
        "domains": [],
        "queryStrategy": ""
      },
      {
        "address": "localhost",
        "domains": [],
        "queryStrategy": ""
      }
    ],
    "tag": "dns"
  },
  "inbounds": [
    {
      "listen": "127.0.0.1",
      "port": 3080,
      "protocol": "socks",
      "settings": {
        "udp": true
      },
      "sniffing": {
        "destOverride": [
          "http",
          "tls",
          "quic"
        ],
        "enabled": true,
        "metadataOnly": false,
        "routeOnly": true
      },
      "tag": "socks-in"
    },
    {
      "listen": "127.0.0.1",
      "port": 3081,
      "protocol": "http",
      "sniffing": {
        "destOverride": [
          "http",
          "tls",
          "quic"
        ],
        "enabled": true,
        "metadataOnly": false,
        "routeOnly": true
      },
      "tag": "http-in"
    }
  ],
  "log": {
    "loglevel": "debug",
	"access": "access.log",
	"error": "error.log"
  },
  "outbounds": [
    {
      "domainStrategy": "AsIs",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "194.67.203.24",
            "port": 443,
            "users": [
              {
                "encryption": "none",
                "flow": "xtls-rprx-vision",
                "id": "ed2ab8fc-6a74-42ce-b136-3bf0255492fa"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "tcp",
        "realitySettings": {
          "fingerprint": "chrome",
          "publicKey": "TrzZNtAHerCUPK7A6OudkGE47P0spcOkRK6NO3w9gg8",
          "serverName": "aws.amazon.com",
          "shortId": "b83c2e00576b253f",
          "spiderX": ""
        },
        "security": "reality"
      },
      "tag": "proxy"
    },
    {
      "domainStrategy": "",
      "protocol": "freedom",
      "tag": "direct"
    },
    {
      "domainStrategy": "",
      "protocol": "freedom",
      "tag": "bypass"
    },
    {
      "protocol": "blackhole",
      "tag": "block"
    },
    {
      "protocol": "dns",
      "proxySettings": {
        "tag": "proxy",
        "transportLayer": true
      },
      "settings": {
        "address": "8.8.8.8",
        "network": "tcp",
        "port": 53,
        "userLevel": 1
      },
      "tag": "dns-out"
    }
  ],
  "policy": {
    "levels": {
      "1": {
        "connIdle": 30
      }
    },
    "system": {
      "statsOutboundDownlink": true,
      "statsOutboundUplink": true
    }
  },
  "routing": {
    "domainStrategy": "AsIs",
    "rules": [
      {
        "inboundTag": [
          "socks-in",
          "http-in"
        ],
        "outboundTag": "dns-out",
        "port": "53",
        "type": "field"
      },
      {
        "outboundTag": "proxy",
        "port": "0-65535",
        "type": "field"
      }
    ]
  },
  "stats": {}
}`

	var config, err = serial.DecodeJSONConfig(bytes.NewReader([]byte(configJson)))
	if err != nil {
		log.Fatal("Ошибка при загрузке конфигурации:", err.Error())
	}
	newConf, err := config.Build()

	if err != nil {
		log.Fatal(err)
	}

	instance, err := core.New(newConf)
	if err != nil {
		log.Fatal("Ошибка при создании инстанса Core:", err)
	}

	if err := instance.Start(); err != nil {
		log.Fatal("Ошибка при запуске Core:", err)
	}

	addr2, err := net2.ParseDestination("tcp:cp.cloudflare.com:80")

	if err != nil {
		log.Fatal(err)
	}

	dial, err := core.Dial(context.Background(), instance, addr2)

	defer func(dial net2.Conn) {
		err := dial.Close()
		if err != nil {
			log.Fatal("close")
		}
	}(dial)

	const size = 1024
	payload := make([]byte, size)

	go func() {
		if _, err := dial.Write(payload); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if _, err := dial.Read(payload); err != nil {
			log.Fatal(err)
		}
	}()

}
