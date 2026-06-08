package socialpost

import (
	"reflect"
	"strings"
	"testing"
)

func hasSocialPostJSONField(t reflect.Type, jsonName string) bool {
	for i := 0; i < t.NumField(); i++ {
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name == jsonName {
			return true
		}
	}
	return false
}

// @Test
func TestTask01SocialPostDTOsExposeAllRequiredFields(t *testing.T) {
	// RegisterPack
	regReqType := reflect.TypeOf(RegisterSocialPostPackRequest{})
	for _, field := range []string{"schema", "workflows", "metrics", "version"} {
		if !hasSocialPostJSONField(regReqType, field) {
			t.Fatalf("RegisterSocialPostPackRequest missing json field %q", field)
		}
	}
	regRespType := reflect.TypeOf(RegisterSocialPostPackResponse{})
	for _, field := range []string{"content_pack_id", "content_type_id", "registered_version"} {
		if !hasSocialPostJSONField(regRespType, field) {
			t.Fatalf("RegisterSocialPostPackResponse missing json field %q", field)
		}
	}

	// Pack status
	statusType := reflect.TypeOf(SocialPostPackStatusResponse{})
	for _, field := range []string{"content_pack_id", "content_type", "schema", "workflows", "metrics", "current_version"} {
		if !hasSocialPostJSONField(statusType, field) {
			t.Fatalf("SocialPostPackStatusResponse missing json field %q", field)
		}
	}

	// Config
	configRespType := reflect.TypeOf(SocialPostConfigResponse{})
	for _, field := range []string{"target_platforms", "default_variant_count", "caption_length_policy", "hashtag_policy", "cover_copy_policy", "tone_style", "forbidden_terms", "config_version"} {
		if !hasSocialPostJSONField(configRespType, field) {
			t.Fatalf("SocialPostConfigResponse missing json field %q", field)
		}
	}
	updateReqType := reflect.TypeOf(UpdateSocialPostConfigRequest{})
	for _, field := range []string{"target_platforms", "default_variant_count", "caption_length_policy", "hashtag_policy", "cover_copy_policy", "tone_style", "forbidden_terms"} {
		if !hasSocialPostJSONField(updateReqType, field) {
			t.Fatalf("UpdateSocialPostConfigRequest missing json field %q", field)
		}
	}
	updateRespType := reflect.TypeOf(UpdateSocialPostConfigResponse{})
	for _, field := range []string{"version_id", "operation_log_id"} {
		if !hasSocialPostJSONField(updateRespType, field) {
			t.Fatalf("UpdateSocialPostConfigResponse missing json field %q", field)
		}
	}

	// Generation run
	createRunReqType := reflect.TypeOf(CreateSocialPostGenerationRunRequest{})
	for _, field := range []string{"topic", "source_content_item_id", "platform", "version_count", "tone_style", "asset_options"} {
		if !hasSocialPostJSONField(createRunReqType, field) {
			t.Fatalf("CreateSocialPostGenerationRunRequest missing json field %q", field)
		}
	}
	createRunRespType := reflect.TypeOf(CreateSocialPostGenerationRunResponse{})
	for _, field := range []string{"generation_run_id", "workflow_run_id", "status"} {
		if !hasSocialPostJSONField(createRunRespType, field) {
			t.Fatalf("CreateSocialPostGenerationRunResponse missing json field %q", field)
		}
	}
	detailType := reflect.TypeOf(SocialPostGenerationRunDetailResponse{})
	for _, field := range []string{"generation_run_id", "workflow_run_id", "status", "content_item_id", "trace", "variants", "error"} {
		if !hasSocialPostJSONField(detailType, field) {
			t.Fatalf("SocialPostGenerationRunDetailResponse missing json field %q", field)
		}
	}
	traceType := reflect.TypeOf(GenerationTrace{})
	for _, field := range []string{"agent_task_ids", "llm_call_log_ids"} {
		if !hasSocialPostJSONField(traceType, field) {
			t.Fatalf("GenerationTrace missing json field %q", field)
		}
	}

	// Variants
	variantRespType := reflect.TypeOf(SocialPostVariantResponse{})
	for _, field := range []string{"id", "content_item_id", "variant_index", "platform", "title", "body", "hashtags", "cover_copy", "tone_style", "status", "content_version_id", "created_at"} {
		if !hasSocialPostJSONField(variantRespType, field) {
			t.Fatalf("SocialPostVariantResponse missing json field %q", field)
		}
	}
	pagedType := reflect.TypeOf(PagedSocialPostVariantsResponse{})
	for _, field := range []string{"items", "pagination"} {
		if !hasSocialPostJSONField(pagedType, field) {
			t.Fatalf("PagedSocialPostVariantsResponse missing json field %q", field)
		}
	}
	selectReqType := reflect.TypeOf(SelectSocialPostVariantRequest{})
	for _, field := range []string{"content_item_id", "note"} {
		if !hasSocialPostJSONField(selectReqType, field) {
			t.Fatalf("SelectSocialPostVariantRequest missing json field %q", field)
		}
	}
	selectRespType := reflect.TypeOf(SelectSocialPostVariantResponse{})
	for _, field := range []string{"selected_variant_id", "content_version_id", "operation_log_id"} {
		if !hasSocialPostJSONField(selectRespType, field) {
			t.Fatalf("SelectSocialPostVariantResponse missing json field %q", field)
		}
	}

	// Assets
	tagsReqType := reflect.TypeOf(GenerateSocialPostTagsRequest{})
	for _, field := range []string{"content_item_id", "variant_id", "platform", "count", "style"} {
		if !hasSocialPostJSONField(tagsReqType, field) {
			t.Fatalf("GenerateSocialPostTagsRequest missing json field %q", field)
		}
	}
	coverReqType := reflect.TypeOf(GenerateSocialPostCoverCopyRequest{})
	for _, field := range []string{"content_item_id", "variant_id", "platform", "count", "style"} {
		if !hasSocialPostJSONField(coverReqType, field) {
			t.Fatalf("GenerateSocialPostCoverCopyRequest missing json field %q", field)
		}
	}
	assetRespType := reflect.TypeOf(GenerateSocialPostAssetResponse{})
	for _, field := range []string{"generation_run_id", "workflow_run_id", "status"} {
		if !hasSocialPostJSONField(assetRespType, field) {
			t.Fatalf("GenerateSocialPostAssetResponse missing json field %q", field)
		}
	}
	assetItemType := reflect.TypeOf(SocialPostAssetItem{})
	for _, field := range []string{"id", "platform", "source_variant_id", "generation_run_id", "result", "created_at"} {
		if !hasSocialPostJSONField(assetItemType, field) {
			t.Fatalf("SocialPostAssetItem missing json field %q", field)
		}
	}
	assetsRespType := reflect.TypeOf(SocialPostAssetsResponse{})
	for _, field := range []string{"tags", "cover_copy", "asset_suggestions", "source_runs"} {
		if !hasSocialPostJSONField(assetsRespType, field) {
			t.Fatalf("SocialPostAssetsResponse missing json field %q", field)
		}
	}
}

// @Test
func TestTask01SocialPostConstantsAndErrorsAreStableContracts(t *testing.T) {
	for _, errValue := range []error{ErrValidation, ErrNotFound, ErrForbidden, ErrConflict, ErrIdempotencyConflict, ErrAgentOutputInvalid, ErrInternal} {
		if errValue == nil {
			t.Fatalf("socialpost domain errors must be declared")
		}
	}
}

// @Test
func TestTask01ServiceInterfaceDeclaresAllSocialPostUseCases(t *testing.T) {
	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, method := range []string{
		"GetPackStatus",
		"RegisterPack",
		"GetConfig",
		"UpdateConfig",
		"CreateGenerationRun",
		"GetGenerationRun",
		"ListVariants",
		"SelectVariant",
		"GenerateTags",
		"GenerateCoverCopy",
		"GetAssets",
	} {
		if _, ok := serviceType.MethodByName(method); !ok {
			t.Fatalf("socialpost Service missing method %s", method)
		}
	}
}

// @Test
func TestTask01StoreInterfaceDeclaresAllSocialPostDataMethods(t *testing.T) {
	storeType := reflect.TypeOf((*Store)(nil)).Elem()
	for _, method := range []string{
		"GetExtensionByProjectID",
		"UpsertExtension",
		"InsertVariant",
		"ListVariants",
		"GetVariantByID",
		"SelectVariantInTx",
		"InsertAsset",
		"ListAssets",
		"InsertOperationLog",
		"CheckIdempotency",
		"StoreIdempotency",
	} {
		if _, ok := storeType.MethodByName(method); !ok {
			t.Fatalf("socialpost Store missing method %s", method)
		}
	}
}

// @Test
func TestTask01NewServiceCreatesDefaultMemoryStore(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService must return non-nil Service")
	}
}

// @Test
func TestTask01NewMemoryStoreCreatesValidStore(t *testing.T) {
	store := NewMemoryStore()
	if store == nil {
		t.Fatal("NewMemoryStore must return non-nil Store")
	}
}

// @Test
func TestTask01ServiceSkeletonMethodsReturnErrInternal(t *testing.T) {
	svc := NewService()
	_, err := svc.GetPackStatus(nil)
	if err == nil {
		t.Fatal("GetPackStatus skeleton must return error")
	}
	_, err = svc.GetConfig(nil, "p-1")
	if err == nil {
		t.Fatal("GetConfig skeleton must return error")
	}
	_, err = svc.GetGenerationRun(nil, "p-1", "gr-1")
	if err == nil {
		t.Fatal("GetGenerationRun skeleton must return error")
	}
	_, err = svc.ListVariants(nil, "p-1", ListSocialPostVariantsRequest{})
	if err == nil {
		t.Fatal("ListVariants skeleton must return error")
	}
	_, err = svc.GetAssets(nil, "p-1", GetSocialPostAssetsRequest{})
	if err == nil {
		t.Fatal("GetAssets skeleton must return error")
	}
}