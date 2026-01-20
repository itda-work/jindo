---
number: 22
title: jd pkg - Claude Code 설정 패키지 매니저
state: done
labels:
  - feat
assignees: []
created_at: '2026-01-20T15:56:23Z'
updated_at: '2026-01-20T15:56:33Z'
closed_at: '2026-01-20T15:56:33Z'
---

## 개요

GitHub의 Claude Code 설정 저장소(agents, commands, skills 등)를 선택적으로 설치하고, 충돌 없이 관리하며, 업데이트를 받을 수 있는 CLI 기능.

## 구현된 명령어

```text
jd pkg
├── repo                    # 저장소 관리
│   ├── add <gh:owner/repo> # 저장소 등록 (git clone)
│   ├── list               # 등록된 저장소 목록
│   ├── remove <namespace>  # 저장소 제거
│   └── update [namespace]  # 저장소 업데이트 (git pull)
├── browse [namespace]      # 저장소 브라우징 (--type 필터)
├── search <query>         # 등록된 저장소에서 검색
├── install <spec>         # 패키지 설치 (namespace:path 형식)
├── uninstall <name>       # 패키지 제거
├── update [name...]       # 업데이트 확인 및 적용 (--apply)
├── list                   # 설치된 패키지 목록
└── info <name>            # 설치된 패키지 상세 정보
```

## 저장 경로

```text
~/.itda-jindo/
├── repos.json           # 등록된 저장소 목록
├── installed.json       # 설치된 패키지 정보
├── repos/               # git clone된 저장소들
│   └── affa-ever/
├── skills/              # 설치된 스킬
├── commands/            # 설치된 커맨드
└── agents/              # 설치된 에이전트
```

## 네임스페이스 규칙

- 자동 생성: `owner 앞4글자` + `-` + `repo 앞4글자`
- 예: `affaan-m/everything-claude-code` → `affa-ever`
- 설치된 패키지: `namespace--name` 형식 (예: `affa-ever--tdd-workflow`)

## 구현된 파일

### 패키지 인프라

- `internal/pkg/git/git.go` - git 유틸리티 (설치 확인/자동 설치, clone, pull 등)
- `internal/pkg/repo/repo.go` - 저장소 관리 (repos.json)
- `internal/pkg/repo/types.go` - 저장소 타입 정의
- `internal/pkg/pkgmgr/pkgmgr.go` - 패키지 설치/제거/업데이트
- `internal/pkg/pkgmgr/types.go` - 패키지 타입 정의

### CLI 명령어

- `internal/cli/pkg.go` - 부모 명령어
- `internal/cli/pkg_repo.go` - repo 서브커맨드 그룹
- `internal/cli/pkg_repo_add.go`
- `internal/cli/pkg_repo_list.go`
- `internal/cli/pkg_repo_remove.go`
- `internal/cli/pkg_repo_update.go`
- `internal/cli/pkg_browse.go`
- `internal/cli/pkg_search.go`
- `internal/cli/pkg_install.go`
- `internal/cli/pkg_uninstall.go`
- `internal/cli/pkg_list.go`
- `internal/cli/pkg_info.go`
- `internal/cli/pkg_update.go`

## 주요 특징

1. **git clone 방식**: GitHub API quota 제한 없음
2. **git 자동 설치**: 미설치 시 사용자 확인 후 자동 설치
3. **네임스페이스 격리**: 로컬 생성 항목과 충돌 방지
4. **오프라인 브라우징**: clone된 저장소에서 패키지 탐색
