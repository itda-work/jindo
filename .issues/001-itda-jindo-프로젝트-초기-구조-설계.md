---
number: 1
title: itda-skills 프로젝트 초기 구조 설계
state: done
labels:
  - epic
assignees: []
created_at: '2026-01-20T10:42:50Z'
updated_at: '2026-01-20T14:22:16Z'
closed_at: '2026-01-20T14:22:16Z'
---

## 개요

Golang 기반 itda-skills CLI와 Claude Code 리소스 구조를 설계하고 구현합니다.

## 목표 구조

```text
itda-skills/
├── cmd/itda-skills/                  # Go CLI 엔트리포인트
├── internal/                      # Go 내부 패키지
├── resources/                     # 임베디드 리소스
│   └── .claude/
│       ├── settings.json
│       ├── commands/itda/
│       ├── skills/
│       ├── agents/itda/
│       └── hooks/itda/
├── install/
│   └── install.sh
├── go.mod
└── README.md
```

## 하위 작업

- [ ] Go 프로젝트 초기화 (go.mod)
- [ ] CLI 기본 구조 (cmd/itda-skills)
- [ ] init 명령 구현
- [ ] .claude/ 리소스 템플릿 생성
- [ ] 설치 스크립트 작성
