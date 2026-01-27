---
number: 45
title: GetWithVendorEnv 메서드 추가 (벤더 환경변수 지원)
state: done
labels:
  - enhancement
assignees: []
created_at: '2026-01-27T05:53:37Z'
updated_at: '2026-01-27T06:22:48Z'
closed_at: '2026-01-27T06:22:48Z'
---

## 배경

skills.tts 등 외부 스킬에서 OpenAI, ElevenLabs 같은 벤더의 기존 환경변수(`OPENAI_API_KEY`, `ELEVENLABS_API_KEY`)도 지원해야 한다.
현재 `GetWithEnv`는 ITDA\_ 프리픽스 환경변수와 config 파일만 조회하므로, 벤더 환경변수를 추가로 조회하는 메서드가 필요하다.

## 변경 내용

`pkg/config/config.go`에 다음 메서드 추가:

```go
// GetWithVendorEnv retrieves a value checking vendor env var, then ITDA_ env var, then config file.
// vendorEnvVar is the vendor-specific environment variable name (e.g., "OPENAI_API_KEY").
func (c *Config) GetWithVendorEnv(key, vendorEnvVar string) (any, bool)
```

조회 순서:

1. 벤더 환경변수 (예: `OPENAI_API_KEY`)
2. ITDA\_ 프리픽스 환경변수 (예: `ITDA_COMMON_API_KEYS_OPENAI`)
3. config 파일 (`common.api_keys.openai`)

## 완료 조건

- `GetWithVendorEnv("common.api_keys.openai", "OPENAI_API_KEY")` 동작 확인
- 벤더 환경변수가 최우선으로 반환됨
- 벤더 환경변수 없으면 기존 `GetWithEnv` 동일하게 동작
- 테스트 추가
