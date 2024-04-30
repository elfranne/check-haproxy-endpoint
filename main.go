package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ruansteve/go-haproxy"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Socket       string
	Backends     bool
	Backend      string
	Servers      bool
	Server       string
	CheckMissing []string
	List         bool
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
			Path:     "socket",
			Env:      "HAPROXY_SOCKET",
			Argument: "socket",
			Default:  "unix:///var/run/haproxy.sock",
			Usage:    "Socket to query for HAProxy stats.",
			Value:    &plugin.Socket,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "backends",
			Env:       "HAPROXY_BACKENDS",
			Argument:  "backends",
			Shorthand: "B",
			Default:   false,
			Usage:     "Check only backends.",
			Value:     &plugin.Backends,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "backend",
			Env:       "HAPROXY_BACKEND",
			Argument:  "backend",
			Shorthand: "b",
			Default:   "",
			Usage:     "Check only specified backend.",
			Value:     &plugin.Backend,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "server",
			Env:       "HAPROXY_SERVER",
			Argument:  "server",
			Shorthand: "s",
			Default:   "",
			Usage:     "Check only specified server.",
			Value:     &plugin.Server,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "servers",
			Env:       "HAPROXY_SERVERS",
			Argument:  "servers",
			Shorthand: "S",
			Default:   false,
			Usage:     "Check only servers.",
			Value:     &plugin.Servers,
		},
		&sensu.SlicePluginConfigOption[string]{
			Path:      "check-missing",
			Env:       "HAPROXY_CHECK-MISSING",
			Argument:  "check-missing",
			Shorthand: "m",
			Default:   []string{},
			Usage:     "Combined with --backends or --servers, will issue an error if there is a missing entry to monitor.",
			Value:     &plugin.CheckMissing,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "list",
			Env:       "HAPROXY_LIST",
			Argument:  "list",
			Shorthand: "l",
			Default:   false,
			Usage:     "Combined with --backends or --servers, list all entries, usefull to debug or generate config",
			Value:     &plugin.List,
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
	if plugin.Backends && plugin.Servers {
		fmt.Print("--backends and --servers are mutually exclusive")
		return sensu.CheckStateUnknown, nil
	}

	return sensu.CheckStateOK, nil
}

func UniqueSliceElements[T comparable](inputSlice []T) []T {
	uniqueSlice := make([]T, 0, len(inputSlice))
	seen := make(map[T]bool, len(inputSlice))
	for _, element := range inputSlice {
		if !seen[element] {
			uniqueSlice = append(uniqueSlice, element)
			seen[element] = true
		}
	}
	return uniqueSlice
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

	List := []string{}

	for _, i := range stats {

		if len(plugin.CheckMissing) > 0 && plugin.Backends && i.SvName == "BACKEND" {
			List = append(List, i.PxName)
		}
		if len(plugin.CheckMissing) > 0 && plugin.Servers && i.SvName != "BACKEND" && i.SvName != "FRONTEND" {
			List = append(List, i.PxName)
		}

		if plugin.List && plugin.Backends && i.SvName == "BACKEND" {
			fmt.Printf("%s\n", i.PxName)
		}
		if plugin.List && plugin.Servers && i.SvName != "BACKEND" && i.SvName != "FRONTEND" {
			fmt.Printf("%s\n", i.SvName)
		}

		if !plugin.List && !plugin.Servers && i.SvName == "BACKEND" && i.Status == "DOWN" {
			if len(plugin.Backend) == 0 || (len(plugin.Backend) > 0 && plugin.Backend == i.PxName) {
				fmt.Printf("Service %s is %s!\n", i.PxName, i.Status)
				Critcount += 1
			}
		}
		if !plugin.List && !plugin.Backends && i.SvName != "BACKEND" && i.Status == "DOWN" {
			if len(plugin.Server) == 0 || (len(plugin.Server) > 0 && plugin.Server == i.SvName) {
				fmt.Printf("Backend %s for service %s is %s!\n", i.SvName, i.PxName, i.Status)
				Warncount += 1
			}
		}
	}

	if len(plugin.CheckMissing) > 0 {
		sort := func(x, y string) bool { return x > y }
		if diff := cmp.Diff(UniqueSliceElements(List), plugin.CheckMissing, cmpopts.SortSlices(sort)); diff != "" {
			fmt.Printf("Missing:\n%s", diff)
			return sensu.CheckStateCritical, nil
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
