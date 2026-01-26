---
number: 18
title: CLI 명령 jindo → jd 변경 및 서브커맨드 단축 별칭 추가
state: done
labels:
  - refactor
assignees: []
created_at: '2026-01-20T13:02:20Z'
updated_at: '2026-01-20T13:04:28Z'
closed_at: '2026-01-20T13:04:28Z'
---

## 변경 사항

### 바이너리명 변경

- `jindo` → `jd`

### 서브커맨드 별칭 추가

- `skills` → `s`
- `commands` → `c`
- `agents` → `a`

### 변경 대상 파일

- cmd/jindo/ → cmd/jd/
- Makefile (BINARY 변수)
- .gitignore
- internal/cli/root.go (Use: "jd")
- internal/cli/version.go
- internal/cli/skills.go (Aliases 추가)
- internal/cli/commands.go (Aliases 추가)

### 모듈명

- `github.com/itda-skills/itda-skills` 유지
