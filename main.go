package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"goimport.moetang.info/nekoq-security/alg/shamir"
	"goimport.moetang.info/nekoq-security/config"
	"goimport.moetang.info/nekoq-security/controller"

	scaffold "github.com/moetang/webapp-scaffold"
)

var preKeyShards []string

func init() {
	b := flag.Bool("genmaster", false, "generate shamir keys")
	premaster := flag.String("preinitmaster", "", "pre-init master key shards. INSECURE")

	flag.Parse()

	if b != nil && *b {
		s, _ := shamir.InitShamirKeys(config.MaxShares, config.MinShares)
		fmt.Println("Key shards generated:")
		for _, v := range s {
			fmt.Println(v)
		}
		os.Exit(0)
	}

	if premaster != nil {
		preKeyShards = strings.Split(*premaster, ",")
	}
}

func main() {
	webscaf, err := scaffold.NewFromConfigFile("nekoq-security.toml")
	if err != nil {
		panic(err)
	}
	c := new(config.NekoQSecurityConfig)
	err = scaffold.ReadCustomConfig("nekoq-security.toml", c)
	if err != nil {
		panic(err)
	}
	err = c.Validate()
	if err != nil {
		panic(err)
	}
	err = c.Init()
	if err != nil {
		panic(err)
	}

	controller.Init(webscaf, c)

	// self-init based on pre-init key shards
	if len(preKeyShards) > 0 {
		go func() {
			init := false
			for _, v := range preKeyShards {
				b := c.FeedShamirKey(v)
				if b {
					init = true
					break
				}
			}
			if !init {
				panic("didn't init nekoq-security")
			} else {
				log.Println("[INFO] self-init nekoq-security done.")
			}
		}()
	}

	err = webscaf.SyncStart()
	if err != nil {
		panic(err)
	}
}
