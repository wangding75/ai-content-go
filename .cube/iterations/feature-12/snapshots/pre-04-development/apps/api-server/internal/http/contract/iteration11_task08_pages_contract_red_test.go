package contract

import (
	"testing"
	"strings"
)

func TestTask08PlatformAdapterAndPluginClientPagesDeclareUIContracts(t *testing.T) {
	adapterPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "platform-adapters", "page.tsx")
	for _, want := range []string{
		"fetchPlatformAdapters",
		"createPlatformAdapter",
		"updatePlatformAdapter",
		"PlatformAdapterResponse",
		"PlatformAdapterDetailResponse",
		"loading",
		"暂无",
		"request_id",
		"page-shell",
		"card",
	} {
		if !strings.Contains(adapterPage, want) {
			t.Fatalf("platform adapters page missing %q", want)
		}
	}

	pluginPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "plugin-clients", "page.tsx")
	for _, want := range []string{
		"fetchPluginClients",
		"createPluginClient",
		"updatePluginClient",
		"rotatePluginClientKey",
		"PluginClientResponse",
		"api_key_once",
		"loading",
		"暂无",
		"request_id",
		"page-shell",
		"card",
	} {
		if !strings.Contains(pluginPage, want) {
			t.Fatalf("plugin clients page missing %q", want)
		}
	}

	nav := readIteration11RepoFile(t, "apps", "web-admin", "app", "global-nav.tsx")
	for _, want := range []string{"/platform-adapters", "平台 Adapter", "/plugin-clients", "插件客户端"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("global nav missing %q", want)
		}
	}
}