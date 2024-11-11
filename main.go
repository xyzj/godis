package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/hdt3213/godis/config"
	"github.com/hdt3213/godis/lib/logger"
	"github.com/hdt3213/godis/lib/utils"
	RedisServer "github.com/hdt3213/godis/redis/server"
	"github.com/hdt3213/godis/tcp"
	"github.com/xyzj/toolbox/gocmd"
)

var banner = `
   ______          ___
  / ____/___  ____/ (_)____
 / / __/ __ \/ __  / / ___/
/ /_/ / /_/ / /_/ / (__  )
\____/\____/\__,_/_/____/

Version: %s
BuildDate: %s
----------------------------------------
`

var defaultProperties = &config.ServerProperties{
	Bind:           "0.0.0.0",
	Port:           6399,
	AppendOnly:     false,
	AppendFilename: "",
	MaxClients:     1000,
	RunID:          utils.RandString(40),
}

var (
	version     string
	builddate   string
	port        = flag.Int("port", 0, "bind port number")
	maxclients  = flag.Int("maxclients", 1000, "max number of clients")
	requirepass = flag.String("requirepass", "", "require auth pass")
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func saveRedisConfig() {
	t := reflect.TypeOf(config.Properties)
	v := reflect.ValueOf(config.Properties)
	// save
	ss := bytes.Buffer{}
	for i := 0; i < t.Elem().NumField(); i++ {
		f := t.Elem().Field(i)
		if key, ok := f.Tag.Lookup("cfg"); ok {
			if key == "runid" ||
				key == "cf" ||
				key == "peers" ||
				key == "self" ||
				key == "masterauth" ||
				key == "cluster-enabled" ||
				key == "cf,omitempty" ||
				strings.HasPrefix(key, "slave-") {
				continue
			}
			ss.WriteString(fmt.Sprintf("%s %+v\n", key, v.Elem().Field(i)))
		}
	}
	os.WriteFile("redis.conf", ss.Bytes(), 0o664)
}

func main() {
	gocmd.DefaultProgram(&gocmd.Info{
		Title:    "godis",
		Descript: "A golang implementation of Redis Server, which intents to provide an example of writing a high concurrent middleware using golang.",
		Ver:      version,
	}).Execute()
	print(fmt.Sprintf(banner, version, builddate))
	logger.Setup(&logger.Settings{
		Path:       "godis-data/log",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})
	configFilename := os.Getenv("CONFIG")
	if configFilename == "" {
		if fileExists("redis.conf") {
			config.SetupConfig("redis.conf")
		} else {
			config.Properties = defaultProperties
		}
	} else {
		config.SetupConfig(configFilename)
	}
	if *port > 0 && *port < 65535 {
		config.Properties.Port = *port
	}
	if *maxclients > 0 {
		config.Properties.MaxClients = *maxclients
	}
	if *requirepass != "" {
		config.Properties.RequirePass = *requirepass
	}
	if config.Properties.Dir == "." {
		config.Properties.Dir = "godis-data"
	}
	if config.Properties.RDBFilename == "test.rdb" {
		config.Properties.RDBFilename = "godis-data/rdb/runtime.rdb"
	}
	saveRedisConfig()
	err := tcp.ListenAndServeWithSignal(&tcp.Config{
		Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
	}, RedisServer.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
