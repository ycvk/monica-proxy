package types

import (
	"context"
	"fmt"
	"monica-proxy/internal/config"
	"monica-proxy/internal/logger"
	"sync/atomic"
	"time"

	lop "github.com/samber/lo/parallel"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	BotChatURL    = "https://api.monica.im/api/custom_bot/chat"
	PreSignURL    = "https://api.monica.im/api/file_object/pre_sign_list_by_module"
	FileUploadURL = "https://api.monica.im/api/files/batch_create_llm_file"
	FileGetURL    = "https://api.monica.im/api/files/batch_get_file"

	// 图片生成相关 API
	ImageGenerateURL = "https://api.monica.im/api/image_tools/text_to_image"
	ImageResultURL   = "https://api.monica.im/api/image_tools/loop_result"
)

// 图片相关常量
const (
	MaxImageSize         = 10 * 1024 * 1024 // 10MB
	ImageModule          = "chat_bot"
	ImageLocation        = "files"
	ImageUploadTimeout   = 30 * time.Second // 图片上传超时时间
	MaxConcurrentUploads = 5                // 最大并发上传数
)

// 支持的图片格式
var SupportedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type ChatGPTRequest struct {
	Model    string        `json:"model"`    // gpt-3.5-turbo, gpt-4, ...
	Messages []ChatMessage `json:"messages"` // 对话数组
	Stream   bool          `json:"stream"`   // 是否流式返回
}

type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content any    `json:"content"` // 可以是字符串或MessageContent数组
}

// MessageContent 消息内容
type MessageContent struct {
	Type     string `json:"type"`                // "text" 或 "image_url"
	Text     string `json:"text,omitempty"`      // 文本内容
	ImageURL string `json:"image_url,omitempty"` // 图片URL
}

// MonicaRequest 为 Monica 自定义 AI 的请求格式
type MonicaRequest struct {
	TaskUID  string    `json:"task_uid"`
	BotUID   string    `json:"bot_uid"`
	Data     DataField `json:"data"`
	Language string    `json:"language"`
	TaskType string    `json:"task_type"`
	ToolData ToolData  `json:"tool_data"`
}

// DataField 在 Monica 的 body 中
type DataField struct {
	ConversationID  string `json:"conversation_id"`
	PreParentItemID string `json:"pre_parent_item_id"`
	Items           []Item `json:"items"`
	TriggerBy       string `json:"trigger_by"`
	UseModel        string `json:"use_model,omitempty"`
	IsIncognito     bool   `json:"is_incognito"`
	UseNewMemory    bool   `json:"use_new_memory"`
}

type Item struct {
	ConversationID string      `json:"conversation_id"`
	ParentItemID   string      `json:"parent_item_id,omitempty"`
	ItemID         string      `json:"item_id"`
	ItemType       string      `json:"item_type"`
	Data           ItemContent `json:"data"`
}

type ItemContent struct {
	Type                   string     `json:"type"`
	Content                string     `json:"content"`
	MaxToken               int        `json:"max_token,omitempty"`
	IsIncognito            bool       `json:"is_incognito,omitempty"` // 是否无痕模式
	FromTaskType           string     `json:"from_task_type,omitempty"`
	ManualWebSearchEnabled bool       `json:"manual_web_search_enabled,omitempty"` // 网页搜索
	UseModel               string     `json:"use_model,omitempty"`
	FileInfos              []FileInfo `json:"file_infos,omitempty"`
}

// ToolData 这里演示放空
type ToolData struct {
	SysSkillList []string `json:"sys_skill_list"`
}

// PreSignRequest 预签名请求
type PreSignRequest struct {
	FilenameList []string `json:"filename_list"`
	Module       string   `json:"module"`
	Location     string   `json:"location"`
	ObjID        string   `json:"obj_id"`
}

// PreSignResponse 预签名响应
type PreSignResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		PreSignURLList []string `json:"pre_sign_url_list"`
		ObjectURLList  []string `json:"object_url_list"`
		CDNURLList     []string `json:"cdn_url_list"`
	} `json:"data"`
}

// MonicaImageRequest 文生图请求结构
type MonicaImageRequest struct {
	TaskUID     string `json:"task_uid"`     // 任务ID
	ImageCount  int    `json:"image_count"`  // 生成图片数量
	Prompt      string `json:"prompt"`       // 提示词
	ModelType   string `json:"model_type"`   // 模型类型，目前只支持 sdxl
	AspectRatio string `json:"aspect_ratio"` // 宽高比，如 1:1, 16:9, 9:16
	TaskType    string `json:"task_type"`    // 任务类型，固定为 text_to_image
}

// FileInfo 文件信息
type FileInfo struct {
	URL        string `json:"url,omitempty"`
	FileURL    string `json:"file_url"`
	FileUID    string `json:"file_uid"`
	Parse      bool   `json:"parse"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	FileType   string `json:"file_type"`
	FileExt    string `json:"file_ext"`
	FileTokens int64  `json:"file_tokens"`
	FileChunks int64  `json:"file_chunks"`
	ObjectURL  string `json:"object_url,omitempty"`
	//Embedding    bool                   `json:"embedding"`
	FileMetaInfo map[string]any `json:"file_meta_info,omitempty"`
	UseFullText  bool           `json:"use_full_text"`
}

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	Data []FileInfo `json:"data"`
}

// FileUploadResponse 文件上传响应
type FileUploadResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items []struct {
			FileName   string `json:"file_name"`
			FileType   string `json:"file_type"`
			FileSize   int64  `json:"file_size"`
			FileUID    string `json:"file_uid"`
			FileTokens int64  `json:"file_tokens"`
			FileChunks int64  `json:"file_chunks"`
			// 其他字段暂时不需要
		} `json:"items"`
	} `json:"data"`
}

// FileBatchGetResponse 获取文件llm处理是否完成
type FileBatchGetResponse struct {
	Data struct {
		Items []struct {
			FileName     string `json:"file_name"`
			FileType     string `json:"file_type"`
			FileSize     int    `json:"file_size"`
			ObjectUrl    string `json:"object_url"`
			Url          string `json:"url"`
			FileMetaInfo struct {
			} `json:"file_meta_info"`
			DriveFileUid  string `json:"drive_file_uid"`
			FileUid       string `json:"file_uid"`
			IndexState    int    `json:"index_state"`
			IndexDesc     string `json:"index_desc"`
			ErrorMessage  string `json:"error_message"`
			FileTokens    int64  `json:"file_tokens"`
			FileChunks    int64  `json:"file_chunks"`
			IndexProgress int    `json:"index_progress"`
		} `json:"items"`
	} `json:"data"`
}

// OpenAIModel represents a model in the OpenAI API format
type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

// OpenAIModelList represents the response format for the /v1/models endpoint
type OpenAIModelList struct {
	Object string        `json:"object"`
	Data   []OpenAIModel `json:"data"`
}

var modelToBotMap = map[string]string{
	"gpt-5":        "gpt_5",
	"gpt-4o":       "gpt_4_o_chat",
	"gpt-4o-mini":  "gpt_4_o_mini_chat",
	"gpt-4.1":      "gpt_4_1",
	"gpt-4.1-mini": "gpt_4_1_mini",
	"gpt-4.1-nano": "gpt_4_1_nano",
	"gpt-4-5":      "gpt_4_5_chat",
	"o3":           "o3",
	"o3-mini":      "openai_o_3_mini",
	"o4-mini":      "o4_mini",

	"claude-haiku-4-5":                  "claude_4_5_haiku",
	"claude-sonnet-4-5":                 "claude_4_5_sonnet",
	"claude-4-sonnet":                   "claude_4_sonnet",
	"claude-4-sonnet-thinking":          "claude_4_sonnet_think",
	"claude-4-opus":                     "claude_4_opus",
	"claude-4-opus-thinking":            "claude_4_opus_think",
	"claude-opus-4-1-20250805-thinking": "claude_4_1_opus_think",
	"claude-3-7-sonnet-thinking":        "claude_3_7_sonnet_think",
	"claude-3-7-sonnet":                 "claude_3_7_sonnet",
	"claude-3-5-haiku":                  "claude_3.5_haiku",

	"gemini-3-pro-preview-thinking": "gemini_3_pro_preview_think",
	"gemini-2.5-pro":                "gemini_2_5_pro",
	"gemini-2.5-flash":              "gemini_2_5_flash",
	"gemini-2.0-flash":              "gemini_2_0",

	"deepseek-v3.1":     "deepseek_v3_1",
	"deepseek-reasoner": "deepseek_reasoner",
	"deepseek-chat":     "deepseek_chat",
	"deepclaude":        "deepclaude",

	"sonar":               "sonar",
	"sonar-reasoning-pro": "sonar_reasoning_pro",

	"grok-3-beta":      "grok_3_beta",
	"grok-4":           "grok_4",
	"grok-code-fast-1": "grok_code_fast_1",
}

func modelToBot(model string) string {
	if botUID, ok := modelToBotMap[model]; ok {
		return botUID
	}
	// 如果未找到映射，则返回原始模型名称
	logger.Warn("未找到模型映射，使用原始名称", zap.String("model", model))
	return model
}

// CustomBotRequest 定义custom bot的请求结构
type CustomBotRequest struct {
	TaskUID        string        `json:"task_uid"`
	BotUID         string        `json:"bot_uid"`
	Data           CustomBotData `json:"data"`
	Language       string        `json:"language"`
	Locale         string        `json:"locale"`
	TaskType       string        `json:"task_type"`
	BotData        BotData       `json:"bot_data"`
	AIRespLanguage string        `json:"ai_resp_language,omitempty"`
}

// CustomBotData custom bot的数据字段
type CustomBotData struct {
	ConversationID      string `json:"conversation_id"`
	Items               []Item `json:"items"`
	PreGeneratedReplyID string `json:"pre_generated_reply_id"`
	PreParentItemID     string `json:"pre_parent_item_id"`
	Origin              string `json:"origin"`
	OriginPageTitle     string `json:"origin_page_title"`
	TriggerBy           string `json:"trigger_by"`
	UseModel            string `json:"use_model"`
	IsIncognito         bool   `json:"is_incognito"`
	UseNewMemory        bool   `json:"use_new_memory"`
	UseMemorySuggestion bool   `json:"use_memory_suggestion"`
}

// BotData bot配置数据
type BotData struct {
	Description    string        `json:"description"`
	LogoURL        string        `json:"logo_url"`
	Name           string        `json:"name"`
	Classification string        `json:"classification"`
	Prompt         string        `json:"prompt"`
	Type           string        `json:"type"`
	UID            string        `json:"uid"`
	ExampleList    []interface{} `json:"example_list"`
	ToolData       BotToolData   `json:"tool_data"`
}

// BotToolData bot工具数据
type BotToolData struct {
	KnowledgeList    []interface{} `json:"knowledge_list"`
	UserSkillList    []interface{} `json:"user_skill_list"`
	SysSkillList     []interface{} `json:"sys_skill_list"`
	UseModel         string        `json:"use_model"`
	ScheduleTaskList []interface{} `json:"schedule_task_list"`
}

// Custom Bot相关的URL
const (
	CustomBotSaveURL    = "https://api.monica.im/api/custom_bot/save_bot"
	CustomBotPublishURL = "https://api.monica.im/api/custom_bot/publish_bot"
	CustomBotPinURL     = "https://api.monica.im/api/custom_bot/pin_bot"
	CustomBotChatURL    = "https://api.monica.im/api/custom_bot/preview_chat"
)

// GetSupportedModels 获取支持的模型列表
func GetSupportedModels() []string {
	models := []string{
		"gpt-5",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-5",
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",

		"claude-sonnet-4-5",
		"claude-4-sonnet",
		"claude-4-sonnet-thinking",
		"claude-4-opus",
		"claude-4-opus-thinking",
		"claude-opus-4-1-20250805-thinking",
		"claude-3-7-sonnet-thinking",
		"claude-3-7-sonnet",
		"claude-3-5-sonnet",
		"claude-3-5-haiku",

		"gemini-2.5-pro",
		"gemini-2.5-flash",
		"gemini-2.0-flash",
		"gemini-1",

		"o1-preview",
		"o3",
		"o3-mini",
		"o4-mini",

		"deepseek-reasoner",
		"deepseek-chat",
		"deepclaude",

		"sonar",
		"sonar-reasoning-pro",

		"grok-3-beta",
		"grok-4",
		"grok-code-fast-1",
	}
	return models
}

// ChatGPTToMonica 将 ChatGPTRequest 转换为 MonicaRequest
func ChatGPTToMonica(cfg *config.Config, chatReq openai.ChatCompletionRequest) (*MonicaRequest, error) {
	if len(chatReq.Messages) == 0 {
		return nil, fmt.Errorf("empty messages")
	}

	// 生成会话ID
	conversationID := fmt.Sprintf("conv:%s", uuid.New().String())

	// 转换消息

	// 设置默认欢迎消息头，不加上就有几率去掉问题最后的十几个token，不清楚是不是bug
	defaultItem := Item{
		ItemID:         fmt.Sprintf("msg:%s", uuid.New().String()),
		ConversationID: conversationID,
		ItemType:       "reply",
		Data:           ItemContent{Type: "text", Content: "__RENDER_BOT_WELCOME_MSG__"},
	}
	var items = make([]Item, 1, len(chatReq.Messages))
	items[0] = defaultItem
	preItemID := defaultItem.ItemID

	for _, msg := range chatReq.Messages {
		if msg.Role == "system" {
			// monica不支持设置prompt，所以直接跳过
			continue
		}
		var msgContext string
		var imgUrl []*openai.ChatMessageImageURL
		if len(msg.MultiContent) > 0 { // 说明应该是多内容，可能是图片内容
			for _, content := range msg.MultiContent {
				switch content.Type {
				case "text":
					msgContext = content.Text
				case "image_url":
					imgUrl = append(imgUrl, content.ImageURL)
				}
			}
		}
		itemID := fmt.Sprintf("msg:%s", uuid.New().String())
		itemType := "question"
		if msg.Role == "assistant" {
			itemType = "reply"
		}

		var content ItemContent
		if len(imgUrl) > 0 {
			// 为图片上传创建带超时的上下文
			uploadCtx, cancel := context.WithTimeout(context.Background(), ImageUploadTimeout)
			defer cancel()

			// 统计上传成功和失败数量
			var successCount, failureCount int64

			// 并发上传图片并收集结果
			uploadResults := lop.Map(imgUrl, func(item *openai.ChatMessageImageURL, _ int) *FileInfo {
				f, err := UploadBase64Image(uploadCtx, cfg, item.URL)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					logger.Error("上传图片失败",
						zap.Error(err),
						zap.String("image_url", item.URL),
						zap.Int("total_images", len(imgUrl)),
					)
					return nil // 返回 nil 表示失败
				}

				if f == nil {
					atomic.AddInt64(&failureCount, 1)
					logger.Warn("图片上传返回空结果", zap.String("image_url", item.URL))
					return nil
				}

				atomic.AddInt64(&successCount, 1)
				return f
			})

			// 过滤掉失败的上传，只保留成功的
			fileIfoList := make([]FileInfo, 0, len(uploadResults))
			for _, result := range uploadResults {
				if result != nil {
					fileIfoList = append(fileIfoList, *result)
				}
			}

			// 记录上传统计信息
			if failureCount > 0 {
				logger.Warn("图片上传完成",
					zap.Int64("success_count", successCount),
					zap.Int64("failure_count", failureCount),
					zap.Int("total_images", len(imgUrl)),
				)
			} else {
				logger.Info("所有图片上传成功",
					zap.Int64("success_count", successCount),
					zap.Int("total_images", len(imgUrl)),
				)
			}

			content = ItemContent{
				Type:        "file_with_text",
				Content:     msgContext,
				FileInfos:   fileIfoList,
				IsIncognito: true,
			}
		} else {
			content = ItemContent{
				Type:        "text",
				Content:     msg.Content,
				IsIncognito: true,
			}
		}

		item := Item{
			ConversationID: conversationID,
			ItemID:         itemID,
			ParentItemID:   preItemID,
			ItemType:       itemType,
			Data:           content,
		}
		items = append(items, item)
		preItemID = itemID
	}

	// 构建请求
	mReq := &MonicaRequest{
		TaskUID: fmt.Sprintf("task:%s", uuid.New().String()),
		BotUID:  modelToBot(chatReq.Model),
		Data: DataField{
			ConversationID:  conversationID,
			Items:           items,
			PreParentItemID: preItemID,
			TriggerBy:       "auto",
			IsIncognito:     true,
			UseModel:        chatReq.Model, //TODO 好像写啥都没影响
			UseNewMemory:    false,
		},
		Language: "auto",
		TaskType: "chat",
	}

	// indent, err := json.MarshalIndent(mReq, "", "  ")
	// if err != nil {
	// 	return nil, err
	// }
	// log.Printf("send: \n%s\n", indent)

	return mReq, nil
}

// ChatGPTToCustomBot 转换ChatGPT请求到Custom Bot请求
func ChatGPTToCustomBot(cfg *config.Config, chatReq openai.ChatCompletionRequest, botUID string) (*CustomBotRequest, error) {
	if len(chatReq.Messages) == 0 {
		return nil, fmt.Errorf("empty messages")
	}

	// 修改customBot请求的模型ID
	chatReq.Model = changeModelToCustomBotModel(chatReq.Model)

	// 生成会话ID
	conversationID := fmt.Sprintf("conv:%s", uuid.New().String())

	// 设置默认欢迎消息
	defaultItem := Item{
		ItemID:         fmt.Sprintf("msg:%s", uuid.New().String()),
		ConversationID: conversationID,
		ItemType:       "reply",
		Data:           ItemContent{Type: "text", Content: "__RENDER_BOT_WELCOME_MSG__"},
	}
	var items = make([]Item, 1, len(chatReq.Messages))
	items[0] = defaultItem
	preItemID := defaultItem.ItemID

	// 提取system消息作为prompt
	var systemPrompt string
	// 转换消息
	for _, msg := range chatReq.Messages {
		if msg.Role == "system" {
			// 将system消息作为prompt
			systemPrompt = msg.Content
			continue
		}

		var msgContext string
		var imgUrl []*openai.ChatMessageImageURL
		if len(msg.MultiContent) > 0 {
			for _, content := range msg.MultiContent {
				switch content.Type {
				case "text":
					msgContext = content.Text
				case "image_url":
					imgUrl = append(imgUrl, content.ImageURL)
				}
			}
		}

		itemID := fmt.Sprintf("msg:%s", uuid.New().String())
		itemType := "question"
		if msg.Role == "assistant" {
			itemType = "reply"
		}

		var content ItemContent
		if len(imgUrl) > 0 {
			// 处理图片上传
			uploadCtx, cancel := context.WithTimeout(context.Background(), ImageUploadTimeout)
			defer cancel()

			var successCount, failureCount int64
			uploadResults := lop.Map(imgUrl, func(item *openai.ChatMessageImageURL, _ int) *FileInfo {
				f, err := UploadBase64Image(uploadCtx, cfg, item.URL)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					logger.Error("上传图片失败",
						zap.Error(err),
						zap.String("image_url", item.URL),
					)
					return nil
				}
				atomic.AddInt64(&successCount, 1)
				return f
			})

			fileIfoList := make([]FileInfo, 0, len(uploadResults))
			for _, result := range uploadResults {
				if result != nil {
					fileIfoList = append(fileIfoList, *result)
				}
			}

			content = ItemContent{
				Type:        "file_with_text",
				Content:     msgContext,
				FileInfos:   fileIfoList,
				IsIncognito: false,
			}
		} else {
			content = ItemContent{
				Type:        "text",
				Content:     msg.Content,
				IsIncognito: false,
			}
		}

		item := Item{
			ConversationID: conversationID,
			ItemID:         itemID,
			ParentItemID:   preItemID,
			ItemType:       itemType,
			Data:           content,
		}
		items = append(items, item)
		preItemID = itemID
	}

	// 生成reply ID
	preGeneratedReplyID := fmt.Sprintf("msg:%s", uuid.New().String())

	// 构建请求
	customBotReq := &CustomBotRequest{
		TaskUID: fmt.Sprintf("task:%s", uuid.New().String()),
		BotUID:  botUID,
		Data: CustomBotData{
			ConversationID:      conversationID,
			Items:               items,
			PreGeneratedReplyID: preGeneratedReplyID,
			PreParentItemID:     preItemID,
			Origin:              fmt.Sprintf("https://monica.im/bots/%s", botUID),
			OriginPageTitle:     "Monica Bot Test",
			TriggerBy:           "auto",
			UseModel:            chatReq.Model, // 使用请求中的模型
			IsIncognito:         false,
			UseNewMemory:        true,
			UseMemorySuggestion: true,
		},
		Language: "auto",
		Locale:   "zh_CN",
		TaskType: "chat",
		BotData: BotData{
			Description:    "Test Bot",
			LogoURL:        "https://assets.monica.im/assets/img/default_bot_icon.jpg",
			Name:           "Test Bot",
			Classification: "custom",
			Prompt:         systemPrompt,
			Type:           "custom_bot",
			UID:            botUID,
			ExampleList:    []interface{}{},
			ToolData: BotToolData{
				KnowledgeList:    []interface{}{},
				UserSkillList:    []interface{}{},
				SysSkillList:     []interface{}{},
				UseModel:         chatReq.Model,
				ScheduleTaskList: []interface{}{},
			},
		},
		AIRespLanguage: "Chinese (Simplified)",
	}

	return customBotReq, nil
}

func changeModelToCustomBotModel(model string) string {
	switch model {
	case "grok-4":
		return "grok-4-0709"
	case "gemini-2.5-pro":
		return "gemini-2.5-pro-thinking"
	default:
		return model
	}
}