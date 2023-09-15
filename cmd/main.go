package main

import (
	"bytes"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	"log"
)

func main() {
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
    "loglevel": "warning"
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

	var config, err = serial.LoadJSONConfig(bytes.NewReader([]byte(configJson)))
	if err != nil {
		log.Fatal("Ошибка при загрузке конфигурации:", err.Error())
	}

	instance, err := core.New(config)
	if err != nil {
		log.Fatal("Ошибка при создании инстанса Core:", err)
	}

	//proxyman.InboundConfig{}

	if err := instance.Start(); err != nil {
		log.Fatal("Ошибка при запуске Core:", err)
	}
}
