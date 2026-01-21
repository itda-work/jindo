---
number: 39
title: 'feat: Default scope를 local 우선으로 변경 및 UX 개선'
state: done
labels:
  - enhancement
assignees: []
created_at: '2026-01-21T17:29:08Z'
updated_at: '2026-01-21T17:30:00Z'
closed_at: '2026-01-21T17:30:00Z'
---

## 개요

로컬 `.claude` 디렉토리가 존재할 경우 기본 scope를 local로 변경하고, 관련 UX를 개선합니다.

## 변경 사항

### 1. Default Scope 로직 변경

- `DefaultScope()` 함수 추가: 로컬 `.claude` 디렉토리가 있으면 local scope 반환
- `ResolveScope()` 함수 추가: global/local 플래그를 처리하고 기본 scope 결정

### 2. 코드 리팩토링

- 모든 명령에서 scope 결정 로직을 `ResolveScope()`로 통일
- 중복 코드 제거 및 일관성 향상

### 3. 명령어 별칭 추가

- `show`: get, view
- `new`: add, create

### 4. list 명령 개선

- hooks도 list 출력에 포함

### 5. 에러 메시지 개선

- 'skill not found' 등의 에러에 scope 정보(local/global) 포함

### 6. 문서 업데이트

- README.md에 Default Scope 섹션 추가
- 각 리소스 경로에 global/local 설명 추가
