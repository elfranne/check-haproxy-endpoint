package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ruansteve/go-haproxy"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Socket string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-haproxy-endpoint ",
			Short:    "Check Haproxy endpoints",
			Keyspace: "sensu.io/plugins/check-haproxy-endpoint/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "socket",
			Env:       "HAPROXY_SOCKET",
			Argument:  "socket",
			Shorthand: "s",
			Default:   "unix:///var/run/haproxy.sock",
			Usage:     "Socket to query for HAProxy stats",
			Value:     &plugin.Socket,
		},
	}
)

func main() {
	useStdin := false
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("Error check stdin: %v\n", err)
		panic(err)
	}
	//Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		log.Println("using stdin")
		useStdin = true
	}

	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, useStdin)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	// TODO
	return sensu.CheckStateOK, nil
}

func executeCheck(event *corev2.Event) (int, error) {
	client := &haproxy.HAProxyClient{
		Addr: plugin.Socket,
	}

	stats, err := client.Stats()
	if err != nil {
		fmt.Printf("could not connect to socket: %s", err)
		return sensu.CheckStateCritical, nil
	}
	Critcount := 0
	Warncount := 0
	for _, i := range stats {
		if i.SvName == "BACKEND" && i.Status == "DOWN" {
			fmt.Printf("Service %s is %s!\n", i.PxName, i.Status)
			Critcount += 1
		}
		if i.SvName != "BACKEND" && i.Status == "DOWN" {
			fmt.Printf("Backend %s for service %s is %s!\n", i.SvName, i.PxName, i.Status)
			Warncount += 1
		}
	}

	if Critcount > 0 {
		return sensu.CheckStateCritical, nil
	}
	if Warncount > 0 {
		return sensu.CheckStateCritical, nil
	}
	fmt.Printf("Haproxy at %s: all systems UP\n", plugin.Socket)
	return sensu.CheckStateOK, nil
}
