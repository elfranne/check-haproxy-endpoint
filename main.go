package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"unsafe"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	URL       string
	AdminUser string
	AdminPass string
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
			Path:      "url",
			Env:       "HAPROXY_URL",
			Argument:  "url",
			Shorthand: "u",
			Default:   "http://demo.haproxy.org/;json",
			Usage:     "URLs to query for HAProxy stats",
			Value:     &plugin.URL,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "admin-user",
			Env:       "HAPROXY_ADMIN_USER",
			Argument:  "admin-user",
			Shorthand: "a",
			Default:   "",
			Usage:     "admin username to be supplied for basic auth, optional",
			Value:     &plugin.AdminUser,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "admin-pass",
			Env:       "HAPROXY_ADMIN_PASS",
			Argument:  "admin-pass",
			Shorthand: "p",
			Default:   "",
			Usage:     "admin password to be supplied for basic auth, optional",
			Value:     &plugin.AdminPass,
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

	jsonData := make(map[string]interface{})

	req, err := http.NewRequest("GET", plugin.URL, nil)
	if err != nil {
		return sensu.CheckStateWarning, fmt.Errorf("failed build request")
	}

	if len(plugin.AdminPass) > 0 && len(plugin.AdminUser) > 0 {
		req.SetBasicAuth(plugin.AdminUser, plugin.AdminPass)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "github.com/elfranne/check-haproxy-endpoint")
	req.Close = true
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return sensu.CheckStateWarning, fmt.Errorf("request failed")
	}
	if response.StatusCode != http.StatusOK {
		return sensu.CheckStateWarning, fmt.Errorf("bad http return code")
	}
	defer response.Body.Close()
	bodyData, _ := io.ReadAll(response.Body)

	json.NewDecoder(response.Body).Decode(&jsonData)

	// debug
	fmt.Printf("http resp: %d\n", response.StatusCode)
	fmt.Printf("body size: %d\n", unsafe.Sizeof(bodyData))
	fmt.Printf("Decoded: %s\n", plugin.URL)
	fmt.Printf("decoded size: %d\n", unsafe.Sizeof(jsonData))
	fmt.Printf("data: %s\n", jsonData)

	for key, metric := range jsonData {
		fmt.Printf("data2: %s, %s\n", metric, key)
	}
	return sensu.CheckStateOK, nil
}
