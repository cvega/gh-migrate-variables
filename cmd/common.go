package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func ConfigureHostname(key string) string {
	hostname := viper.GetString(key)
	if hostname != "" {
		hostname = strings.TrimPrefix(hostname, "http://")
		hostname = strings.TrimPrefix(hostname, "https://")
		hostname = strings.TrimSuffix(hostname, "/api/v3")
		hostname = strings.TrimSuffix(hostname, "/")
		hostname = fmt.Sprintf("https://%s/api/v3", hostname)
		viper.Set(key, hostname)
	}

	fmt.Printf("\n%s\n", getHostnameMessage(hostname))

	httpProxy := viper.GetString("HTTP_PROXY")
	httpsProxy := viper.GetString("HTTPS_PROXY")

	fmt.Printf("🔄 Proxy: %s\n\n", getProxyStatus(httpProxy, httpsProxy))

	return hostname
}

func getHostnameMessage(hostname string) string {
	if hostname != "" {
		return fmt.Sprintf("🔗 Using: GitHub Enterprise Server: %s", hostname)
	}
	return "📡 Using: GitHub.com"
}

func getProxyStatus(httpProxy, httpsProxy string) string {
	if httpProxy != "" || httpsProxy != "" {
		return "✅ Configured"
	}
	return "❌ Not configured"
}
