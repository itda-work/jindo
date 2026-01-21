---
number: 33
title: 'feat: adapt command - AI-based skill/hook/agent customization'
state: done
labels:
  - feature
assignees: []
created_at: '2026-01-21T14:03:08Z'
updated_at: '2026-01-21T14:12:43Z'
closed_at: '2026-01-21T14:12:43Z'
---

## 개요

스킬, 훅, 에이전트를 설치 후 각자의 워크플로우와 상황에 맞게 AI 대화로 맞춤화하고, 변경 이력을 관리하는 시스템.

## 배경

- 다른 사람의 스킬/훅/명령을 그대로 복사하면 각자의 상황에 맞지 않는 경우가 많음
- 수동으로 수정하려면 어디를 어떻게 바꿔야 하는지 파악하기 어려움
- 원본 업데이트 시 사용자 커스터마이징이 덮어써질 위험

## 구현 내용

### 1. adapt 명령

```bash
jd skills adapt <skill-id> [--global|--local]
jd hooks adapt <hook-name> [--global|--local]
jd agents adapt <agent-id> [--global|--local]
```

**동작 흐름:**

1. 시작 전 현재 버전 `.history/`에 백업
2. `claude` 명령 실행 (맞춤화 프롬프트 전달)
3. AI가 사용자 상황 질문 → 스킬 수정 → 완료 안내
4. 대화 종료 후 변경 감지 → manifest.json 업데이트

**자동완성:** 기존 `skillNameCompletion`, `hookNameCompletion`, `agentNameCompletion` 재사용

### 2. 버전 관리 명령

```bash
jd skills history <id>           # 변경 이력 조회
jd skills revert <id> [version]  # 특정 버전으로 롤백
jd hooks history <name>
jd hooks revert <name>
jd agents history <id>
jd agents revert <id>
```

**저장 구조:**

```text
~/.claude/skills/my-skill/
├── skill.md
└── .history/
    ├── v001-2026-01-21T10-30-00.md
    ├── v002-2026-01-21T14-20-00.md
    └── manifest.json
```

### 3. 프롬프트 관리 명령

```bash
jd prompts list                  # 프롬프트 목록
jd prompts show <name>           # 내용 보기
jd prompts edit <name>           # 편집
jd prompts reset <name>          # embed 기본값으로 초기화
```

**프롬프트 저장:**

- 기본값: Go embed로 바이너리에 내장
- 오버라이드: `~/.claude/jindo/prompts/adapt-skill.md` 등
- 오버라이드 파일 있으면 우선 사용, 없으면 embed 사용

**adapt 실행 시 Tip 출력:**

```text
💡 Tip: 맞춤화 프롬프트를 수정하려면: jd prompts edit adapt-skill
```

## 구현 파일

| 파일                             | 역할                                    |
| -------------------------------- | --------------------------------------- |
| `internal/cli/skills_adapt.go`   | adapt 명령                              |
| `internal/cli/skills_history.go` | history 명령                            |
| `internal/cli/skills_revert.go`  | revert 명령                             |
| `internal/cli/prompts.go`        | prompts 부모 명령                       |
| `internal/cli/prompts_*.go`      | list, show, edit, reset                 |
| `internal/prompt/prompt.go`      | 프롬프트 로드 로직 (embed + 오버라이드) |
| `internal/prompt/embed.go`       | embed 기본 프롬프트                     |
| `internal/skill/history.go`      | 버전 관리 로직                          |
| `internal/hook/history.go`       | hooks용 버전 관리                       |
| `internal/agent/history.go`      | agents용 버전 관리                      |

## 비목표 (Non-goals)

- 대화 중 중간 버전 관리 (시작 vs 최종만 관리)
- TUI 기반 대화형 질문 (Claude Code 대화로 처리)
- 제작자가 맞춤화 포인트를 미리 정의하는 스키마
