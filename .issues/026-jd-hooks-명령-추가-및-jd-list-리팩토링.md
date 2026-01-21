---
number: 26
title: jd hooks 명령 추가 및 jd list 리팩토링
state: done
labels:
  - enhancement
assignees: []
created_at: '2026-01-21T03:02:36Z'
updated_at: '2026-01-21T03:07:57Z'
closed_at: '2026-01-21T03:07:57Z'
---

## Background

jindo는 Claude Code의 skills, agents, commands를 관리하는 CLI 도구입니다. 현재 hooks 관련 명령이 없어서 Claude Code의 hook 기능을 관리할 수 없는 상태입니다.

## Problem

1. **hooks 관리 부재**: Claude Code hooks 설정을 CLI로 관리할 수 없음
2. **list 중복 로직**: jd list가 jd skills list, jd agents list 등과 별개로 store 접근 로직을 가짐

## Goal

### 1. jd hooks 명령 추가 (통합 wizard 스타일)

- jd hooks new: matcher, event type, command 입력받아 settings.json에 등록 + 선택적 스크립트 파일 생성
- jd hooks list: settings.json의 hook 규칙 목록 표시
- jd hooks show <name>: 특정 hook 상세 정보
- jd hooks edit <name>: hook 수정
- jd hooks delete <name>: hook 삭제

### 2. jd list 리팩토링

- 기존 runSkillsList, runAgentsList, runCommandsList 함수를 내부적으로 호출
- 각 타입에 맞는 출력 형식 유지

## Non-goals

- Hook 스크립트 디버깅/테스트 기능
- Hook 실행 로그 관리

## Constraints

- 기존 skills/agents/commands와 동일한 CLI 패턴 유지
- settings.json 파싱/수정은 안전하게 (기존 설정 보존)
