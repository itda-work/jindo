---
number: 16
title: YAML frontmatter 파싱 개선
state: wip
labels:
  - enhancement
assignees: []
created_at: '2026-01-20T10:47:19Z'
updated_at: '2026-01-20T14:03:36Z'
---

description에 특수 문자(<, >, :)나 이스케이프된 줄바꿈(\n)이 포함된 경우 YAML 파싱 실패. 더 관대한 파싱 또는 정규식 기반 추출로 개선 필요
