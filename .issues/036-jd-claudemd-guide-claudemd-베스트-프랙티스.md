---
number: 36
title: 'feat: jd claudemd guide - CLAUDE.md 베스트 프랙티스 안내'
state: open
labels:
  - feature
assignees: []
created_at: '2026-01-21T17:20:41Z'
updated_at: '2026-01-21T17:20:41Z'
---

## 개요

CLAUDE.md 작성에 대한 베스트 프랙티스를 AI 기반으로 안내하는 명령 추가

## 기능

- gemini CLI를 활용하여 CLAUDE.md 작성 가이드 제공
- 현재 CLAUDE.md 분석 후 개선점 제안
- 스타일별 템플릿 및 예시 제공

## 사용 예시

```bash
jd claudemd guide                   # 일반 가이드 출력
jd claudemd guide --analyze         # 현재 파일 분석 후 개선점 제안
jd claudemd guide --template        # 템플릿 출력
```

## 제약 사항

- gemini CLI가 설치되어 있어야 함 (optional - 없으면 기본 가이드만 표시)
