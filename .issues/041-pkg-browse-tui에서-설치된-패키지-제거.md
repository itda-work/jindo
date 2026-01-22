---
number: 41
title: 'feat: pkg browse TUI에서 설치된 패키지 제거 기능 추가'
state: done
labels:
  - feature
assignees: []
created_at: '2026-01-22T14:11:15Z'
updated_at: '2026-01-22T14:21:35Z'
closed_at: '2026-01-22T14:21:35Z'
---

## Problem 1-Pager: pkg browse TUI 패키지 제거 기능

## Background

jindo의 `jd pkg browse` TUI는 현재 패키지를 탐색하고 설치하는 기능을 제공합니다:

- space로 패키지 선택
- enter로 선택된 패키지 설치
- 설치된 패키지는 [✓] 표시

하지만 설치된 패키지를 제거하려면:

1. TUI를 종료 (q)
2. `jd pkg uninstall <name>` 명령 실행
3. 다시 browse TUI 진입

이는 워크플로우를 중단시키고 불편합니다.

## Problem

TUI 내에서 설치된 패키지를 제거할 수 없음:

- 설치는 TUI에서 가능하지만, 제거는 별도 CLI 명령으로만 가능
- 패키지를 잘못 설치했거나 테스트 후 제거하고 싶을 때 TUI를 나가야 함
- 대칭적이지 않은 UX (install은 TUI, uninstall은 CLI)

## Goal

TUI 내에서 'd' 키를 눌러 설치된 패키지를 제거할 수 있게 함.

성공 기준:

- [ ] 설치된 패키지([✓] 표시)에서 'd' 키를 누르면 제거 확인 프롬프트 표시
- [ ] 확인 후 `manager.Uninstall()` 호출하여 패키지 제거
- [ ] 제거 성공 시 TUI 상태 업데이트 (IsInstalled = false, [✓] 제거)
- [ ] 제거 실패 시 에러 메시지 표시
- [ ] 설치되지 않은 패키지에서는 'd' 키가 무시됨
- [ ] help 텍스트에 'd' 키 설명 추가

## UX 디자인

**키 바인딩**: `d` (delete)

- vim-style 삭제 키, 개발자에게 직관적
- 커서가 설치된 패키지([✓])에 있을 때만 동작

**확인 프롬프트**:

- "Uninstall 'affa-ever--web-fetch'? [y/N]" 형태의 메시지 표시
- y/Y 입력 시 제거 진행
- 다른 입력 시 취소

**상태 업데이트**:

- 제거 성공 → message: "✓ Uninstalled affa-ever--web-fetch"
- 제거 실패 → message: "✗ Failed to uninstall: [error message]"
- 패키지 아이템의 IsInstalled 플래그 업데이트

## Non-goals

- 다중 선택 후 배치 제거 (현재는 단일 패키지만)
- 확인 없이 즉시 제거 (안전성을 위해 항상 확인)
- 제거 후 되돌리기(undo) 기능

## Implementation Plan

### 1. Key binding 추가 (internal/tui/browse.go)

```go
Uninstall: key.NewBinding(
    key.WithKeys("d"),
    key.WithHelp("d", "uninstall"),
),
```

### 2. Update() 메서드에 케이스 추가

```go
case key.Matches(msg, keys.Uninstall):
    item := m.getCurrentItem()
    if item != nil && item.IsInstalled {
        // Show confirmation and uninstall
        return m, m.confirmUninstall()
    }
    return m, nil
```

### 3. 확인 및 제거 로직

- 확인 프롬프트 표시를 위한 상태 추가 또는 메시지 처리
- manager.Uninstall() 호출
- 결과에 따른 상태 업데이트

### 4. Help 텍스트 업데이트

```go
help := "↑/↓: navigate  ←/→/tab: switch tab  space: select  a: select all  enter: install  d: uninstall  q: quit"
```

## Constraints

- 기존 Bubble Tea TUI 구조 유지
- 현재 설치/업데이트 워크플로우와 일관성 유지
- pkgmgr.Manager의 Uninstall() 메서드 사용
- 에러 처리 패턴 통일

---

Clarified on: 2026-01-22
