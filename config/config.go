package config

import (
	"log"

	"github.com/spf13/viper"
)

var PORT string

var MY_URL string
var MY_PUBLIC_KEY string

var IS_PRO bool
var INITIAL_TARGET string
var INITIAL_BLOCK_REWARD int64

var MAX_PEERS int
var POTENTIAL_PEERS []string

func init() {
	viper.SetConfigName("iitkbucks-config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("[WARNING] Unable to locate configuration file")
	}

	viper.AutomaticEnv()

	PORT = viper.GetString("port")

	MY_URL = viper.GetString("myUrl")
	MY_PUBLIC_KEY = viper.GetString("myPublicKey")

	IS_PRO = viper.GetBool("isPro")
	INITIAL_TARGET = viper.GetString("initialTarget")
	INITIAL_BLOCK_REWARD = viper.GetInt64("initialBlockReward")

	MAX_PEERS = viper.GetInt("maxPeers")
	POTENTIAL_PEERS = viper.GetStringSlice("potentialPeers")
}
