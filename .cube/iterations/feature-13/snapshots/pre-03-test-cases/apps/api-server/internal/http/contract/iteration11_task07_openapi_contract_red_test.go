package contract

import (
	"testing"
	"strings"
)

func TestTask07OpenAPIAndWebClientDeclareIteration11Contracts(t *testing.T) {
	openapi := readIteration11RepoFile(t, "openapi", "openapi.yaml")
	for _, want := range []string{
		"/api/v1/platform-adapters:",
		"operationId: createPlatformAdapter",
		"operationId: listPlatformAdapters",
		"operationId: getPlatformAdapter",
		"operationId: updatePlatformAdapter",
		"/api/v1/plugin-clients:",
		"operationId: registerPluginClient",
		"operationId: listPluginClients",
		"operationId: updatePluginClient",
		"operationId: rotatePluginClientKey",
		"/api/v1/plugin-auth/token:",
		"operationId: authenticatePlugin",
		"/api/v1/plugin/publish-jobs:",
		"operationId: listPluginPublishJobs",
		"operationId: lockPluginPublishJob",
		"operationId: markPluginPublishJobFilled",
		"operationId: markPluginPublishJobPublished",
		"operationId: markPluginPublishJobFailed",
		"/api/v1/platform-collect-logs:",
		"operationId: submitPlatformCollectLog",
		"operationId: listPlatformCollectLogs",
		"operationId: getPlatformCollectLog",
		"operationId: confirmPlatformCollectLogMetrics",
		"operationId: rotateCallbackToken",
		"operationId: updateCallbackAuth",
		"operationId: receiveExternalCallback",
		"operationId: listCallbackLogs",
		"operationId: testExternalCallback",
		"Idempotency-Key",
		"X-External-Binding-Id",
		"IDEMPOTENCY_CONFLICT",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("iteration11 openapi missing %q", want)
		}
	}

	apiClient := readIteration11RepoFile(t, "apps", "web-admin", "lib", "api.ts")
	for _, want := range []string{
		"export type PlatformAdapterResponse",
		"export type PlatformAdapterDetailResponse",
		"export type PluginClientResponse",
		"export type PluginPublishJobResponse",
		"export type PlatformCollectLogResponse",
		"export type PlatformCollectLogDetailResponse",
		"export type ExternalCallbackLogResponse",
		"export async function createPlatformAdapter",
		"export async function fetchPlatformAdapters",
		"export async function fetchPlatformAdapter",
		"export async function updatePlatformAdapter",
		"export async function createPluginClient",
		"export async function fetchPluginClients",
		"export async function updatePluginClient",
		"export async function rotatePluginClientKey",
		"export async function authenticatePlugin",
		"export async function fetchPluginPublishJobs",
		"export async function lockPluginPublishJob",
		"export async function markPluginPublishJobFilled",
		"export async function markPluginPublishJobPublished",
		"export async function markPluginPublishJobFailed",
		"export async function submitPlatformCollectLog",
		"export async function fetchPlatformCollectLogs",
		"export async function fetchPlatformCollectLog",
		"export async function confirmPlatformCollectLogMetrics",
		"export async function rotateCallbackToken",
		"export async function updateCallbackAuth",
		"export async function receiveExternalCallback",
		"export async function fetchCallbackLogs",
		"export async function testExternalCallback",
		"Idempotency-Key",
	} {
		if !strings.Contains(apiClient, want) {
			t.Fatalf("iteration11 web api client missing %q", want)
		}
	}
	if strings.Contains(apiClient, "export type PlatformAdapterResponse = any") || strings.Contains(apiClient, "export type PluginClientResponse = any") {
		t.Fatalf("iteration11 web api client must not use any for public response types")
	}
}