---
number: 23
title: jd pkg browse TUI 인터랙티브 모드
state: done
labels:
  - feat
  - tui
assignees: []
created_at: '2026-01-21T00:55:00Z'
updated_at: '2026-01-20T16:22:05Z'
closed_at: '2026-01-20T16:22:05Z'
---

## Background

현재 `jd pkg browse affa-ever`처럼 네임스페이스를 지정해야만 패키지 목록을 볼 수 있음. 등록된 모든 저장소의 패키지를 한눈에 보고, 선택적으로 설치하는 인터랙티브 경험이 필요함.

## Problem

- 여러 저장소를 등록한 경우 각각 browse 명령을 실행해야 함
- 설치 여부, 업데이트 가능 여부를 한눈에 파악하기 어려움
- 설치할 패키지를 선택하려면 별도로 install 명령을 입력해야 함

## Goal

`jd pkg browse` (인자 없이) 실행 시 TUI가 열리고:

1. 등록된 모든 저장소의 패키지를 탭별로 탐색
2. 설치 상태, 업데이트 가능 여부 시각적 표시
3. 체크박스로 다중 선택 후 일괄 설치

### 성공 기준

- TUI에서 Skills, Commands, Agents, Hooks 탭 전환
- 각 항목에 설치 상태 표시 (✓ 설치됨, ↑ 업데이트 가능, 빈칸)
- Space로 다중 선택, Enter로 일괄 설치
- q 또는 Esc로 종료

## Non-goals

- ~~패키지 상세 내용 미리보기 (향후 구현)~~ → 구현 완료
- 설치된 패키지 삭제 기능 (기존 uninstall 명령 사용)
- 저장소 추가/삭제 (기존 repo 명령 사용)

## Constraints

- TUI 라이브러리: Bubble Tea (Charm)
- 탭 네비게이션: 상단 탭 바, 좌우 방향키 또는 Tab으로 전환
- 선택 방식: 체크박스 다중 선택 후 Enter로 일괄 설치
- hooks 탭: 완전히 구현 (browse, install, uninstall 포함)

## UI 스케치

```text
┌─ jd pkg browse ──────────────────────────────────────┐
│ [Skills]  Commands   Agents   Hooks                  │
├──────────────────────────────────────────────────────┤
│ affa-ever                                            │
│   [ ] security-review                                │
│   [✓] tdd-workflow                         installed │
│                                                      │
│ anth-clau                                            │
│   [ ] web-fetch                                      │
│   [↑] api-helper                      update available│
│                                                      │
├──────────────────────────────────────────────────────┤
│ Space: select  Enter: install  Tab: switch tab  q: quit │
└──────────────────────────────────────────────────────┘
```

---

Clarified on: 2026-01-21

## 작업 내역

### 1차 구현 (3a8a021)

- `internal/tui/browse.go`: Bubble Tea 기반 TUI 모델 구현
- Skills, Commands, Agents, Hooks 4개 탭 네비게이션
- 설치 상태 표시 (`[✓]` installed, `[*]` selected, `[ ]` available)
- 다중 선택 및 일괄 설치 기능
- hook 패키지 타입 지원 추가 (`repo/types.go`, `repo/repo.go`, `pkgmgr/pkgmgr.go`)
- `jd pkg browse` 인자 없이 실행 시 TUI 모드 자동 실행

### 2차 구현 - 미리보기 패널 (873196f)

- 30:70 비율의 좌우 분할 레이아웃 적용
- 왼쪽 패널: 패키지 목록
- 오른쪽 패널: 선택된 패키지의 파일 내용 미리보기 (최대 30줄)
- 커서 이동 시 실시간 미리보기 업데이트
- skill의 경우 `SKILL.md` 파일 로드, command/agent는 해당 `.md` 파일 로드

### 키바인딩

| 키                          | 기능             |
| --------------------------- | ---------------- |
| `↑/↓` 또는 `j/k`            | 목록 탐색        |
| `←/→` 또는 `h/l` 또는 `Tab` | 탭 전환          |
| `Space`                     | 항목 선택/해제   |
| `a`                         | 전체 선택/해제   |
| `Enter`                     | 선택한 항목 설치 |
| `q` 또는 `Esc`              | 종료             |

### 파일 구조

```text
internal/
├── tui/
│   └── browse.go          # TUI 모델, 뷰, 키바인딩
├── cli/
│   └── pkg_browse.go      # TUI 진입점 (인자 없으면 TUI 실행)
└── pkg/
    ├── repo/
    │   ├── types.go       # TypeHook 추가
    │   └── repo.go        # scanHooks() 추가
    └── pkgmgr/
        └── pkgmgr.go      # installHook() 추가
```
