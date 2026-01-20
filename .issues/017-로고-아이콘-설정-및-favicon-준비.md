---
number: 17
title: 로고 아이콘 설정 및 favicon 준비
state: wip
labels:
  - enhancement
assignees: []
created_at: '2026-01-20T11:09:34Z'
updated_at: '2026-01-20T14:12:01Z'
---

## 개요

`~/Downloads/itda-jindo.png` 로고 아이콘을 프로젝트에 추가하고, README.md 상단 배치 및 favicon 풀 세트를 준비합니다.

## 작업 내용

### 1. 로고 이미지 복사

- 원본 로고를 `assets/logo.png`로 복사

### 2. README.md 로고 배치

- README.md 최상단에 로고 이미지 삽입
- 적절한 크기로 조정 (예: 128px 또는 200px)

### 3. Favicon 풀 세트 생성

다양한 플랫폼/브라우저 지원을 위한 파일 생성:

- `favicon.ico` - 멀티 사이즈 ICO (16x16, 32x32, 48x48)
- `favicon-16x16.png`
- `favicon-32x32.png`
- `apple-touch-icon.png` (180x180) - iOS Safari
- `android-chrome-192x192.png` - Android Chrome
- `android-chrome-512x512.png` - Android Chrome (고해상도)
- `mstile-150x150.png` - Windows 타일

### 4. 저장 위치

모든 이미지 파일은 `assets/` 디렉토리에 저장

## 참고

- 원본 이미지: 진돗개 캐릭터 로고 (투명 배경 PNG)
- favicon 생성 시 ImageMagick 또는 온라인 도구 활용 가능
