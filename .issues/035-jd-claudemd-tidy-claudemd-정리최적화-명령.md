---
number: 35
title: 'feat: jd claudemd tidy - CLAUDE.md 정리/최적화 명령'
state: open
labels:
  - feature
assignees: []
created_at: '2026-01-21T17:20:40Z'
updated_at: '2026-01-21T17:20:40Z'
---

## 개요

복잡해진 CLAUDE.md 파일을 분석하고 정리/최적화하는 명령 추가

## 기능

- gemini CLI를 활용하여 CLAUDE.md 분석 및 개선
- 중복 지침 제거
- 구조 개선 및 일관성 확보
- 여러 스타일 옵션 제공:
  - `minimal`: 핵심 지침만 간결하게
  - `detailed`: 상세한 설명 포함
  - `structured`: 섹션별 명확한 구조화

## 사용 예시

```bash
jd claudemd tidy                    # 현재 CLAUDE.md 정리
jd claudemd tidy --style minimal    # minimal 스타일로 정리
jd claudemd tidy --dry-run          # 변경 사항 미리보기
jd claudemd tidy --global           # 글로벌 CLAUDE.md 정리
```

## 제약 사항

- gemini CLI가 설치되어 있어야 함 (없으면 에러 메시지 표시)
- 원본 파일 백업 후 정리 수행
