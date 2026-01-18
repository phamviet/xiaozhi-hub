package seeds

import (
	"log"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/core"
)

func Seed(app core.App) error {
	if err := seedSysParams(app); err != nil {
		return err
	}

	if err := seedModelProviders(app); err != nil {
		return err
	}

	if err := seedModelConfigs(app); err != nil {
		return err
	}

	return nil
}

func seedSysParams(app core.App) error {
	log.Println("Seeding sys_params...")

	// 1. memory.system_prompt: use current database value or default if empty
	prompt := ""
	existingPrompt, err := app.FindFirstRecordByData("sys_params", "name", "memory.system_prompt")
	if err == nil {
		prompt = existingPrompt.GetString("value")
	}
	if prompt == "" {
		prompt = "You are a memory assistant. Your goal is to update the long-term memory of an AI agent based on the latest conversation. You will be provided with \"Previous Memories\" (if any) and the \"Current Conversation\". Combine the information from both to create a new, concise, and updated summary of the agent's knowledge about the user, their preferences, and important facts. Maintain a consistent persona for the agent and ensure that the summary remains useful for future interactions."
	}

	params := []map[string]string{
		{"name": "memory.system_prompt", "value": prompt},
		{"name": "server.secret", "value": uuid.New().String()},
		{"name": "server.websocket", "value": "ws://REPLACE_WITH_YOUR_SERVER_IP:8090/xiaozhi/v1"},
		{"name": "server.ota", "value": "http://REPLACE_WITH_YOUR_SERVER_IP:8090/xiaozhi/ota/"},
	}

	collection, err := app.FindCollectionByNameOrId("sys_params")
	if err != nil {
		return err
	}

	for _, p := range params {
		existing, err := app.FindFirstRecordByData("sys_params", "name", p["name"])
		if err != nil {
			record := core.NewRecord(collection)
			record.Set("name", p["name"])
			record.Set("value", p["value"])
			if err := app.Save(record); err != nil {
				return err
			}
			log.Printf("Created sys_param: %s\n", p["name"])
		} else {
			// Only update server.secret if it's empty
			if p["name"] == "server.secret" && existing.GetString("value") == "" {
				existing.Set("value", p["value"])
				if err := app.Save(existing); err != nil {
					return err
				}
				log.Printf("Updated empty server.secret\n")
			}
			log.Printf("Skipped sys_param: %s (already exists)\n", p["name"])
		}
	}

	return nil
}

func seedModelProviders(app core.App) error {
	log.Println("Seeding model_providers...")
	collection, err := app.FindCollectionByNameOrId("model_providers")
	if err != nil {
		return err
	}

	providers := []map[string]interface{}{
		{"id": "1xie8bsxclw5hlg", "name": "OPENAI", "provider_code": "openai", "model_type": "ASR"},
		{"id": "740an4pex5vb8lc", "name": "OPENAPI", "provider_code": "openai", "model_type": "LLM"},
		{"id": "d0bkxbo8cz1yh72", "name": "OPENAI", "provider_code": "openai", "model_type": "TTS"},
		{"id": "e5hl2el2sstm5af", "name": "EdgeTTS", "provider_code": "edge", "model_type": "TTS"},
		{"id": "brzsnbmj7f1uoff", "name": "Gemini", "provider_code": "gemini", "model_type": "LLM"},
		{"id": "i0k1rmawtjuoai6", "name": "VAD_SileroVAD", "provider_code": "silero", "model_type": "VAD"},
		{"id": "9n4cc3xz3r6m7j0", "name": "Intent_LLM", "provider_code": "intent_llm", "model_type": "Intent"},
		{"id": "j4urro5wwnfj1z6", "name": "MemLocalShort", "provider_code": "mem_local_short", "model_type": "Memory"},
	}

	for _, p := range providers {
		existing, err := app.FindRecordById("model_providers", p["id"].(string))
		if err != nil {
			record := core.NewRecord(collection)
			record.Set("id", p["id"].(string))
			record.Set("name", p["name"])
			record.Set("provider_code", p["provider_code"])
			record.Set("model_type", p["model_type"])
			if err := app.Save(record); err != nil {
				return err
			}
			log.Printf("Created model_provider: %s\n", p["name"])
		} else {
			log.Printf("Skipped model_provider: %s (already exists)\n", p["name"])
			_ = existing
		}
	}
	return nil
}

func seedModelConfigs(app core.App) error {
	log.Println("Seeding model_config...")
	collection, err := app.FindCollectionByNameOrId("model_config")
	if err != nil {
		return err
	}

	configs := []map[string]interface{}{
		{
			"id":          "jyrpnlnlzo1iw5b",
			"model_name":  "google/gemini-2.5-flash-lite",
			"model_type":  "LLM",
			"is_default":  true,
			"is_enabled":  true,
			"provider_id": "740an4pex5vb8lc",
			"config_json": map[string]interface{}{
				"model_name": "google/gemini-2.5-flash-lite",
				"base_url":   "https://ai-gateway.vercel.sh/v1",
			},
		},
		{
			"id":          "4yikhnheajdkpca",
			"model_name":  "google/gemini-2.5-flash",
			"model_type":  "LLM",
			"is_default":  false,
			"is_enabled":  true,
			"provider_id": "740an4pex5vb8lc",
			"config_json": map[string]interface{}{
				"model_name": "google/gemini-2.5-flash",
				"base_url":   "https://ai-gateway.vercel.sh/v1",
			},
		},
		{
			"id":          "5spyb7djzk30w5m",
			"model_name":  "EdgeTTS",
			"model_type":  "TTS",
			"is_default":  true,
			"is_enabled":  true,
			"provider_id": "e5hl2el2sstm5af",
			"config_json": map[string]interface{}{
				"private_voice": "vi-VN-NamMinhNeural",
			},
		},
		{
			"id":          "ayt0xgt2dn75ueo",
			"model_name":  "SileroVAD",
			"model_type":  "VAD",
			"is_default":  true,
			"is_enabled":  true,
			"provider_id": "i0k1rmawtjuoai6",
			"config_json": nil,
		},
		{
			"id":          "a5qr4fqmsgyynuy",
			"model_name":  "OPENAI",
			"model_type":  "ASR",
			"is_default":  true,
			"is_enabled":  true,
			"provider_id": "1xie8bsxclw5hlg",
			"config_json": map[string]interface{}{
				"base_url":   "http://192.168.1.34:4700/v1/audio/transcriptions",
				"model_name": "gpt-4o-mini-transcribe",
				"output_dir": "tmp/",
			},
		},
		{
			"id":          "1o49isbt4oha4rg",
			"model_name":  "Short Memory",
			"model_type":  "Memory",
			"is_default":  true,
			"is_enabled":  true,
			"provider_id": "j4urro5wwnfj1z6",
			"config_json": map[string]interface{}{
				"llm": "4yikhnheajdkpca",
			},
		},
		{
			"id":          "rebh8m5grt2hkhf",
			"model_name":  "google/gemini-2.5-flash-lite",
			"model_type":  "Intent",
			"is_default":  true,
			"is_enabled":  true,
			"provider_id": "9n4cc3xz3r6m7j0",
			"config_json": map[string]interface{}{
				"llm":  "jyrpnlnlzo1iw5b",
				"type": "intent_llm",
			},
		},
	}

	for _, c := range configs {
		existing, err := app.FindRecordById("model_config", c["id"].(string))
		if err != nil {
			record := core.NewRecord(collection)
			record.Set("id", c["id"].(string))
			record.Set("model_name", c["model_name"])
			record.Set("model_type", c["model_type"])
			record.Set("is_default", c["is_default"])
			record.Set("is_enabled", c["is_enabled"])
			record.Set("provider_id", c["provider_id"])
			record.Set("config_json", c["config_json"])
			if err := app.Save(record); err != nil {
				return err
			}
			log.Printf("Created model_config: %s\n", c["model_name"])
		} else {
			log.Printf("Skipped model_config: %s (already exists)\n", c["model_name"])
			_ = existing
		}
	}

	return nil
}
