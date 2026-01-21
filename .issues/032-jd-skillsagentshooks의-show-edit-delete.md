---
number: 32
title: jd skills/agents/hooks의 show, edit, delete 명령에 ID 자동완성 지원
state: done
labels:
  - enhancement
assignees: []
created_at: '2026-01-21T06:10:11Z'
updated_at: '2026-01-21T06:13:10Z'
closed_at: '2026-01-21T06:13:10Z'
---

## Background

`jd skills show <id>`, `jd agents delete <id>` 등의 명령에서 사용자가 id를 직접 타이핑해야 함.
존재하는 skill/agent/hook 이름을 모르면 먼저 `list` 명령으로 확인해야 하는 불편함이 있음.

## Problem

`show`, `edit`, `delete` 서브커맨드에서 첫 번째 인자(id)에 대한 쉘 자동완성이 지원되지 않음. Tab 키로 기존 리소스 이름을 자동완성할 수 없음.

## Goal

- `jd skills show <Tab>` → 존재하는 skill 이름들 제안
- `jd agents edit <Tab>` → 존재하는 agent 이름들 제안
- `jd hooks delete <Tab>` → 존재하는 hook 이름들 제안
- `--local` 플래그가 있으면 local scope에서, 없으면 global scope에서 자동완성
- 대상: skills, agents, hooks 각각의 show, edit, delete 명령 (총 9개 서브커맨드)

## Non-goals

- `new` 명령에는 적용하지 않음 (새로 생성하는 것이므로)
- global + local 합쳐서 제안하지 않음 (scope별로 분리)

## Constraints

- cobra의 `ValidArgsFunction`을 사용하여 구현
- 기존 `hooks_new.go`의 completion 패턴 참고
- 플래그 파싱 시점에 `--local`/`--global` 값을 읽어야 함
