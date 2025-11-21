package types

import (
	"testing"
)

// TestGetSupportedModels 测试获取支持的模型列表
func TestGetSupportedModels(t *testing.T) {
	models := GetSupportedModels()

	// 验证模型列表不为空
	if len(models) == 0 {
		t.Fatal("支持的模型列表不应为空")
	}

	// 验证模型列表长度应该等于 modelToBotMap 的长度
	if len(models) != len(modelToBotMap) {
		t.Errorf("模型列表长度 (%d) 不等于 modelToBotMap 长度 (%d)", len(models), len(modelToBotMap))
	}

	// 验证每个返回的模型都在 modelToBotMap 中
	for _, model := range models {
		if _, exists := modelToBotMap[model]; !exists {
			t.Errorf("模型 %s 在 GetSupportedModels 中返回但不在 modelToBotMap 中", model)
		}
	}

	// 验证 modelToBotMap 中的每个模型都能被 GetSupportedModels 返回
	modelSet := make(map[string]bool)
	for _, model := range models {
		modelSet[model] = true
	}

	for model := range modelToBotMap {
		if !modelSet[model] {
			t.Errorf("模型 %s 在 modelToBotMap 中但未被 GetSupportedModels 返回", model)
		}
	}
}

// TestModelToBot 测试模型到 Bot UID 的映射
func TestModelToBot(t *testing.T) {
	testCases := []struct {
		model    string
		expected string
	}{
		{"gpt-4o", "gpt_4_o_chat"},
		{"claude-sonnet-4-5", "claude_4_5_sonnet"},
		{"claude-3-5-sonnet", "claude_3.5_sonnet"},
		{"gemini-2.5-pro", "gemini_2_5_pro"},
		{"o1-preview", "o1_preview"},
		{"deepseek-v3.1", "deepseek_v3_1"},
		{"grok-4", "grok_4"},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			result := modelToBot(tc.model)
			if result != tc.expected {
				t.Errorf("modelToBot(%s) = %s; 期望 %s", tc.model, result, tc.expected)
			}
		})
	}
}

// TestModelToBotFallback 测试未知模型的 fallback 行为
func TestModelToBotFallback(t *testing.T) {
	unknownModel := "unknown-model-xyz"
	result := modelToBot(unknownModel)

	// fallback 应该返回原始模型名称
	if result != unknownModel {
		t.Errorf("modelToBot(%s) = %s; 期望 fallback 到原始名称 %s", unknownModel, result, unknownModel)
	}
}

// TestAllSupportedModelsHaveMapping 测试所有支持的模型都有有效的映射
func TestAllSupportedModelsHaveMapping(t *testing.T) {
	models := GetSupportedModels()

	for _, model := range models {
		botUID := modelToBot(model)

		// 验证 botUID 不为空
		if botUID == "" {
			t.Errorf("模型 %s 映射到空的 botUID", model)
		}

		// 验证映射结果与 modelToBotMap 中的值一致
		expectedBotUID, exists := modelToBotMap[model]
		if !exists {
			t.Errorf("模型 %s 不在 modelToBotMap 中", model)
			continue
		}

		if botUID != expectedBotUID {
			t.Errorf("模型 %s: modelToBot 返回 %s，但 modelToBotMap 中是 %s", model, botUID, expectedBotUID)
		}
	}
}

// TestNoHardcodedModels 测试确保没有硬编码的模型列表与 map 不一致
func TestNoHardcodedModels(t *testing.T) {
	// 这个测试确保 GetSupportedModels 完全依赖 modelToBotMap
	// 通过比较两者的长度来验证
	models := GetSupportedModels()

	if len(models) != len(modelToBotMap) {
		t.Errorf("检测到硬编码问题：GetSupportedModels 返回 %d 个模型，但 modelToBotMap 有 %d 个条目",
			len(models), len(modelToBotMap))
	}

	// 额外验证：确保所有模型都来自 modelToBotMap
	for _, model := range models {
		if _, exists := modelToBotMap[model]; !exists {
			t.Errorf("发现硬编码模型 %s，它不在 modelToBotMap 中", model)
		}
	}
}
