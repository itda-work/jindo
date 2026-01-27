---
number: 46
title: DefaultTemplate에 TTS 섹션 추가
state: done
labels:
  - enhancement
assignees: []
created_at: '2026-01-27T05:53:44Z'
updated_at: '2026-01-27T06:51:09Z'
closed_at: '2026-01-27T06:51:09Z'
---

## 배경

skills.tts에서 jindo 통합 설정을 사용하려면 DefaultTemplate에 OpenAI/ElevenLabs API 키와 TTS 설정 섹션이 포함되어야 한다.

## 변경 내용

`pkg/config/config.go`의 `DefaultTemplate` 상수에 다음 섹션 추가:

```toml
[common.api_keys]
# openai = "your-api-key"
# elevenlabs = "your-api-key"

# [skills.tts]
# default_provider = "openai"
# default_voice = "alloy"
```

## 완료 조건

- `jd config init` 실행 시 생성되는 템플릿에 openai/elevenlabs 키와 skills.tts 섹션 포함
- 기존 tiingo/polygon 키는 유지
