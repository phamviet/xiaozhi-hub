# PocketBase Database Schema Documentation

This document provides an overview of the database collections and their fields in the Xiaozhi Hub project.

## Table of Contents
- [users](#users)
- [ai_device](#ai_device)
- [model_providers](#model_providers)
- [model_config](#model_config)
- [ai_agent](#ai_agent)
- [ai_agent_chat_history](#ai_agent_chat_history)
- [ai_agent_template](#ai_agent_template)
- [sys_params](#sys_params)
- [user_credentials](#user_credentials)
- [sys_config](#sys_config)

---

## users
Auth collection for managing system users.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| username | text | Yes | |
| email | email | Yes | |
| emailVisibility | bool | No | |
| verified | bool | No | |
| name | text | No | |
| avatar | file | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## ai_device
Stores information about AI devices connected to the hub.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key (MAC-like pattern) |
| user | relation | No | Relates to `users` |
| mac_address | text | Yes | MAC address pattern |
| agent | relation | No | Relates to `ai_agent` |
| last_connected | date | No | |
| board | text | No | |
| auto_update | bool | No | |
| firmware_version | text | No | |
| attributes | json | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## model_providers
Defines different AI model providers (e.g., OpenAI, Anthropic, etc.).

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| name | text | Yes | Presentable |
| model_type | select | No | ASR, LLM, Intent, Memory, TTS, VAD, Plugin, RAG, VLLM |
| provider_code | text | Yes | |
| fields | json | No | |
| sort | number | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## model_config
Specific configurations for AI models.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| model_name | text | Yes | |
| model_type | select | Yes | ASR, LLM, TTS, VAD, Memory, Intent |
| provider_id | relation | No | Relates to `model_providers` |
| is_default | bool | No | |
| is_enabled | bool | No | |
| config_json | json | No | |
| credential | relation | No | Relates to `user_credentials` |
| remark | text | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## ai_agent
Configured AI agents with specific personalities and model settings.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| user | relation | No | Relates to `users` |
| agent_name | text | Yes | |
| lang_code | text | No | ISO 639-1 code |
| system_prompt | text | No | |
| summary_memory | text | No | |
| asr_model_id | text | No | |
| vad_model_id | text | No | |
| llm_model_id | text | No | |
| vllm_model_id | text | No | |
| tts_model_id | text | No | |
| tts_voice_id | text | No | |
| mem_model_id | text | No | |
| intent_model_id | text | No | |
| chat_history_enabled | bool | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## ai_agent_chat_history
Logs of conversations between users/devices and AI agents.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| mac_address | text | No | |
| agent_id | text | Yes | |
| conversation_id | text | No | |
| content | text | No | |
| chat_type | select | No | Values: 1, 2 |
| audio | file | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## ai_agent_template
Templates for creating new AI agents.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| name | text | No | Presentable |
| lang_code | text | No | |
| system_prompt | text | No | |
| asr_model_id | text | No | |
| vad_model_id | text | No | |
| llm_model_id | text | No | |
| vllm_model_id | text | No | |
| tts_model_id | text | No | |
| tts_voice_id | text | No | |
| mem_model_id | text | No | |
| intent_model_id | text | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## sys_params
System-wide parameters/settings.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| name | text | Yes | |
| value | text | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## user_credentials
API keys and other credentials for model providers.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| name | text | Yes | Presentable |
| api_key | text | Yes | |
| api_url | url | No | |
| remark | text | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |

## sys_config
System configuration stored in JSON format.

| Field | Type | Required | Options |
|-------|------|----------|---------|
| id | text | Yes | Primary Key |
| name | text | Yes | |
| value | json | No | |
| disabled | bool | No | |
| remark | text | No | |
| created | autodate | Yes | |
| updated | autodate | Yes | |
