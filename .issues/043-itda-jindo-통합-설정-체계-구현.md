---
number: 43
title: 'feat: itda-skills 통합 설정 체계 구현'
state: done
labels:
  - feature
assignees: []
created_at: '2026-01-24T15:10:05Z'
updated_at: '2026-01-24T15:29:05Z'
closed_at: '2026-01-24T15:29:05Z'
---

## Background

itda-skills 하위에 여러 유틸리티(skills.quant-data, skills.igm, skills.hangul, skills.web-auto)가
존재하며, 각각이 독립적인 설정 파일을 가질 경우 관리 복잡성이 증가합니다.
현재는 문제가 없지만, 유틸리티가 늘어날수록 설정 파일 난립, API 키 중복 관리,
설정 방식 불일치 문제가 예상됩니다.

## Problem

- `~/.config/` 아래 qd, igm, hangul 등 폴더가 난립할 가능성
- 동일 API 키를 여러 설정 파일에 중복 저장해야 함
- 각 유틸리티가 다른 설정 방식/구조를 사용하면 사용자 경험 저하
- **예방적 설계**로 미리 통합 체계를 수립할 필요

## Goal

1. **단일 설정 파일**: `~/.config/itda-skills/config.toml` 하나로 모든 설정 통합
2. **계층형 구조**: 공통 설정 + 유틸리티별 섹션으로 구분
3. **CLI 통합**: 모든 유틸리티가 동일한 설정 명령어 체계 사용
4. **구현 위치**: `jindo` 앱에서 통합 설정 체계 구현
5. **활용 가이드**: 다른 skills 명령들이 이 체계를 활용하는 방법 문서화

### 예상 설정 파일 구조

```toml
# ~/.config/itda-skills/config.toml

[common]
default_market = "kr"

[common.api_keys]
tiingo = "xxx"
polygon = "yyy"

[skills.quant-data]
default_format = "json"

[skills.quant-data.sources.krx]
delay = 1000

[skills.igm]
# skills.igm 전용 설정

[skills.hangul]
# skills.hangul 전용 설정

[skills.web-auto]
# skills.web-auto 전용 설정
```

## Non-goals

- 외부 저장소(DB, 클라우드) 동기화
- API 키 암호화 저장 (평문 TOML)
- GUI 기반 설정 도구
- 기존 설정 파일 자동 마이그레이션

## Constraints

- **언어**: Go
- **포맷**: TOML
- **경로**: XDG Base Directory 스펙 준수 (`~/.config/itda-skills/`)

## 작업 내용

### 1. 설정 라이브러리 구현 (pkg/config/)

- [x] TOML 파서 연동 (pelletier/go-toml/v2)
- [x] 설정 파일 경로 관리 (~/.config/itda-skills/config.toml)
- [x] 환경변수 오버라이드 지원 (JINDO\_\* 형식)
- [x] 공통 설정 및 앱별 설정 읽기/쓰기 API

### 2. CLI 명령어 구현 (jd config)

- [x] `jd config init` - 설정 파일 초기화
- [x] `jd config set <key> <value>` - 설정값 변경
- [x] `jd config get <key>` - 설정값 조회
- [x] `jd config list` - 전체 설정 출력
- [x] `jd config edit` - 에디터로 설정 파일 열기
- [x] `jd config guide` - 활용 가이드 출력

### 3. 활용 가이드 문서화

- [x] 다른 skills에서 통합 설정을 활용하는 방법 문서화 (`jd config guide`)
- [x] Go 모듈로 import하여 사용하는 예제 코드

## 구현 결과

### 생성된 파일

| 경로                           | 설명                             |
| ------------------------------ | -------------------------------- |
| `pkg/config/paths.go`          | XDG 경로 관리                    |
| `pkg/config/dotnotation.go`    | 점 표기법 키 파싱                |
| `pkg/config/config.go`         | Config 타입 및 Load/Save/Get/Set |
| `internal/cli/config.go`       | 부모 명령                        |
| `internal/cli/config_init.go`  | init 명령                        |
| `internal/cli/config_set.go`   | set 명령                         |
| `internal/cli/config_get.go`   | get 명령                         |
| `internal/cli/config_list.go`  | list 명령                        |
| `internal/cli/config_edit.go`  | edit 명령                        |
| `internal/cli/config_guide.go` | guide 명령                       |

### 다른 skills에서 사용하기

```go
import "github.com/itda-skills/jindo/pkg/config"

cfg, err := config.Load()
if err != nil {
    return err
}

// 환경변수 우선, 없으면 config 파일에서 읽기
apiKey, found := cfg.GetWithEnv("common.api_keys.tiingo")
```

## 명령어 예시

```bash
# 설정 초기화
jd config init

# API 키 설정
jd config set common.api_keys.tiingo "YOUR_API_KEY"

# 앱별 설정
jd config set skills.quant-data.default_format "table"

# 설정 조회
jd config get common.api_keys.tiingo
jd config list

# 에디터로 직접 편집
jd config edit
```
