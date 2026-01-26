---
number: 20
title: 'feat: update 명령 추가 (zap update와 동일한 옵션 지원)'
state: done
labels:
  - feat
assignees: []
created_at: '2026-01-20T15:03:33Z'
updated_at: '2026-01-20T15:07:14Z'
closed_at: '2026-01-20T15:07:14Z'
---

## 개요

zap의 update 명령과 동일한 기능의 update 명령을 jd CLI에 추가합니다.

## 옵션

- `--check, -c`: 업데이트 확인만 (설치하지 않음)
- `--force, -f`: 확인 없이 업데이트
- `--yes, -y`: 확인 없이 업데이트 (`--force`와 동일)
- `--version, -v <version>`: 특정 버전으로 업데이트
- `--script`: OS 설치 스크립트로 업데이트 (curl/PowerShell)
- 별칭: `up`

## 구현 사항

1. `internal/updater` 패키지 구현 (GitHub Releases 연동)
2. `internal/cli/update.go` 작성
3. 기능:
   - GitHub Releases에서 최신 버전 확인
   - 현재 버전과 비교 후 자동 업데이트
   - 개발 빌드(`dev`) 감지 및 안내
   - 체크섬 검증, 권한 확인
   - 진행 상황 표시

## 참고

- GitHub 저장소: itda-skills/itda-skills
- 설치 스크립트: jsdelivr CDN 사용
