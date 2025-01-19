package core

import (
	"flag"
	"fmt"
	"node/core/internal"
	"node/global"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Viper(path ...string) *viper.Viper {
	var config string

	// search config path: command -> env -> default
	if len(path) == 0 {
		flag.StringVar(&config, "c", "", "choose a config file")
		flag.Parse()
		if config == "" {
			if configEnv := os.Getenv(internal.ConfigEnv); configEnv == "" {
				switch gin.Mode() {
				case gin.DebugMode:
					config = internal.ConfigDebugFile
					break
				case gin.ReleaseMode:
					config = internal.ConfigReleaseFile
					break
				case gin.TestMode:
					config = internal.ConfigTestFile
					break
				}
			} else {
				config = configEnv
			}
		}
	} else {
		config = path[0]
	}

	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed: ", e.Name)
		if err = v.Unmarshal(&global.NODE_CONFIG); err != nil {
			fmt.Println(err)
		}
	})
	if err = v.Unmarshal(&global.NODE_CONFIG); err != nil {
		fmt.Println(err)
	}

	return v
}
