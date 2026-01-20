---
number: 15
title: skills list JSON 출력에서 실제 파일 경로 표시
state: wip
labels:
  - enhancement
assignees: []
created_at: '2026-01-20T10:47:18Z'
updated_at: '2026-01-20T13:57:15Z'
---

현재 macOS 파일시스템의 대소문자 무관 특성으로 인해 skill.md로 요청해도 SKILL.md가 열리는데,
JSON path 필드에는 실제 파일명(SKILL.md)이 아닌 요청 경로(skill.md)가 표시됨.
실제 파일명을 정확히 표시하도록 개선
