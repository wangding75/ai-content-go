package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestTask08CharactersPageSupportsRoleFilterCreateLoadingEmptyErrorSuccess(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/projects/[projectId]/novel/characters/page.tsx")
	if err != nil {
		t.Fatalf("read characters page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"ProjectWorkspaceNav",
		"fetchCharacters",
		"createCharacter",
		"角色筛选",
		"新增人物",
		"Profile JSON",
		"创建成功",
		"加载态",
		"空状态：暂无人物",
		"role=\"alert\"",
		"role=\"status\"",
		"request_id",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected characters page to contain %s", required)
		}
	}
}
