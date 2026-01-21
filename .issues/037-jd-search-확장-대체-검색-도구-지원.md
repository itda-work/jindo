---
number: 37
title: 'feat: jd search 확장 - 대체 검색 도구 지원 (Tavily, Exa)'
state: open
labels:
  - feature
assignees: []
created_at: '2026-01-21T17:20:43Z'
updated_at: '2026-01-21T17:20:43Z'
---

## 개요

Claude 자체 검색이 느린 경우를 위해 대체 검색 도구를 지원하는 기능 추가

## 배경

- Claude의 내장 WebSearch 도구가 느리거나 제한적인 경우가 있음
- Tavily, Exa 등 전문 검색 API가 더 빠르고 정확한 결과를 제공할 수 있음

## 기능

1. **대체 검색 도구 설정**
   - settings.json에 기본 검색 도구 지정 가능
   - Tavily, Exa 등 지원

2. **API 키 관리**
   - 환경 변수 또는 settings.json에서 API 키 관리

3. **검색 명령 확장**
   - `jd search web <query>` 형태로 웹 검색 실행

## 사용 예시

```bash
# 검색 도구 설정
jd config set search.provider tavily
jd config set search.api_key $TAVILY_API_KEY

# 웹 검색 실행
jd search web "Claude Code best practices"
jd search web "Tavily API documentation" --provider exa
```

## 조사 필요

- Tavily API 스펙 및 가격 정책
- Exa API 스펙 및 가격 정책
- 기타 대안 검색 도구 조사

## 제약 사항

- 외부 API 사용 시 해당 서비스의 API 키 필요
- API 키 없으면 해당 기능 비활성화
