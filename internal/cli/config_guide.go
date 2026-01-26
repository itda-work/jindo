package cli

import (
	"io"

	"github.com/spf13/cobra"
)

var configGuideCmd = &cobra.Command{
	Use:   "guide",
	Short: "Show guide for using config in other skills",
	Long:  `Display a guide explaining how to use the unified configuration system in other skills.`,
	RunE:  runConfigGuide,
}

func init() {
	configCmd.AddCommand(configGuideCmd)
}

func runConfigGuide(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	_, _ = io.WriteString(cmd.OutOrStdout(), configGuideContent)
	return nil
}

const configGuideContent = `# itda-skills 통합 설정 활용 가이드

## 개요

itda-skills는 ~/.config/itda-skills/config.toml에서 모든 skills의 설정을 통합 관리합니다.
각 skill은 jindo의 config 패키지를 import하여 설정을 읽고 쓸 수 있습니다.

## 설정 파일 구조

` + "```toml" + `
# ~/.config/itda-skills/config.toml

[common]
default_market = "kr"

[common.api_keys]
tiingo = "your-api-key"
polygon = "your-api-key"

[skills.quant-data]
default_format = "json"

[skills.quant-data.sources.krx]
delay = 1000

[skills.igm]
# igm 전용 설정

[skills.hangul]
# hangul 전용 설정
` + "```" + `

## Go 코드에서 사용하기

### 1. import

` + "```go" + `
import "github.com/itda-skills/jindo/pkg/config"
` + "```" + `

### 2. 설정 로드

` + "```go" + `
cfg, err := config.Load()
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
` + "```" + `

### 3. 값 읽기 (점 표기법)

` + "```go" + `
// 단일 값 읽기
format, err := cfg.Get("skills.quant-data.default_format")
if err != nil {
    // 키가 없거나 에러 발생
}
fmt.Println(format) // "json"

// 타입 단언 필요
if formatStr, ok := format.(string); ok {
    // formatStr 사용
}
` + "```" + `

### 4. 환경변수 오버라이드 지원

` + "```go" + `
// GetWithEnv는 환경변수를 먼저 확인합니다
// JINDO_SKILLS_QUANT_DATA_DEFAULT_FORMAT 환경변수가 있으면 그 값을 반환
value, found := cfg.GetWithEnv("skills.quant-data.default_format")
if !found {
    // 설정이 없음
}
` + "```" + `

환경변수 형식: JINDO_<KEY> (대문자, 점은 밑줄로 변환)
- skills.quant-data.default_format → JINDO_SKILLS_QUANT_DATA_DEFAULT_FORMAT
- common.api_keys.tiingo → JINDO_COMMON_API_KEYS_TIINGO

### 5. 값 쓰기

` + "```go" + `
// 값 설정 (중간 경로가 없으면 자동 생성)
err := cfg.Set("skills.quant-data.cache_ttl", 3600)
if err != nil {
    return err
}

// 저장
err = cfg.Save()
if err != nil {
    return err
}
` + "```" + `

### 6. API 키 읽기 헬퍼 예제

` + "```go" + `
func GetAPIKey(name string) (string, error) {
    cfg, err := config.Load()
    if err != nil {
        return "", err
    }

    key, found := cfg.GetWithEnv("common.api_keys." + name)
    if !found {
        return "", fmt.Errorf("API key not configured: %s", name)
    }

    keyStr, ok := key.(string)
    if !ok {
        return "", fmt.Errorf("API key is not a string: %s", name)
    }

    return keyStr, nil
}

// 사용 예:
// tiingoKey, err := GetAPIKey("tiingo")
` + "```" + `

## 권장 사항

### 키 네이밍 규칙

- skill 전용 설정: skills.<skill-name>.<key>
- 공통 API 키: common.api_keys.<provider>
- 공통 설정: common.<key>

### 타입 처리

config.Get()은 any를 반환하므로 타입 단언이 필요합니다:
- 문자열: value.(string)
- 정수: value.(int64)
- 실수: value.(float64)
- 불리언: value.(bool)
- 맵: value.(map[string]any)

### 기본값 처리

` + "```go" + `
func getFormat(cfg *config.Config) string {
    val, err := cfg.Get("skills.quant-data.default_format")
    if err != nil {
        return "json" // 기본값
    }
    if s, ok := val.(string); ok {
        return s
    }
    return "json" // 기본값
}
` + "```" + `

## CLI 명령어

` + "```bash" + `
# 설정 초기화
jd config init

# 값 설정
jd config set skills.quant-data.default_format table

# 값 조회
jd config get skills.quant-data.default_format

# 전체 설정 출력
jd config list

# 에디터로 편집
jd config edit
` + "```" + `
`
