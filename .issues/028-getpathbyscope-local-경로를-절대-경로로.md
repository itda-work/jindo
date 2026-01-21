---
number: 28
title: GetPathByScope Local 경로를 절대 경로로 반환
state: done
labels:
  - bug
assignees: []
created_at: '2026-01-21T04:46:02Z'
updated_at: '2026-01-21T06:21:08Z'
closed_at: '2026-01-21T06:21:08Z'
---

## 문제

GetPathByScope()가 Local scope일 때 상대 경로(".claude/skills")를 반환합니다.

```go
case ScopeLocal:
    return filepath.Join(localClaudeDir, subdir)  // ".claude/skills" (상대 경로)
```

Store의 expandDir()가 ~ 확장만 처리하므로, CWD가 변경되면 잘못된 경로를 참조할 수 있습니다.

## 해결 방안

절대 경로로 반환:

```go
case ScopeLocal:
    cwd, err := os.Getwd()
    if err != nil {
        return GetGlobalPath(subdir) // fallback
    }
    return filepath.Join(cwd, localClaudeDir, subdir)
```
