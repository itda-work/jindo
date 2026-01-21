---
number: 38
title: 'feat: jd guide codex/gemini - 외부 AI 도구 활용 가이드'
state: open
labels:
  - feature
assignees: []
created_at: '2026-01-21T17:20:45Z'
updated_at: '2026-01-21T17:20:45Z'
---

## 개요

codex CLI, gemini CLI 등 외부 AI 도구를 효과적으로 활용하는 방법을 안내하는 가이드 명령 추가

## 기능

1. **codex CLI 가이드** (`jd guide codex`)
   - codex CLI 설치 방법
   - 주요 사용법 및 옵션
   - Claude Code와 함께 활용하는 패턴
   - 유용한 사용 시나리오

2. **gemini CLI 가이드** (`jd guide gemini`)
   - gemini CLI 설치 방법
   - 주요 사용법 및 옵션
   - Claude Code와 함께 활용하는 패턴
   - MCP 서버 연동 방법

## 사용 예시

```bash
jd guide codex              # codex CLI 활용 가이드
jd guide gemini             # gemini CLI 활용 가이드
jd guide codex --install    # 설치 방법만 출력
jd guide gemini --examples  # 예시 중심 출력
```

## 조사 필요

- codex CLI 최신 기능 및 사용법
- gemini CLI 최신 기능 및 사용법
- 두 도구의 강점과 적합한 사용 케이스

## 기존 guide 명령과 통합

- 현재 `jd guide skills`, `jd guide commands` 등이 있음
- 동일한 패턴으로 `jd guide codex`, `jd guide gemini` 추가
