package parser

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileParser_parseLine(t *testing.T) {
	// Given: 파서 인스턴스 생성
	parser := NewFileParser("")

	testCases := []struct {
		name             string
		line             string
		expectError      bool
		expectedEmail    string
		expectedPhone    string
		expectedCreditUp bool
	}{
		{
			name:             "신용점수 상승 사용자 정상 파싱",
			line:             "Duser780641_29@example.fake 000-0420-2932 Y",
			expectError:      false,
			expectedEmail:    "Duser780641_29@example.fake",
			expectedPhone:    "000-0420-2932",
			expectedCreditUp: true,
		},
		{
			name:             "짧은 이메일 정상 파싱",
			line:             "Duser1_1@example.fake 000-6320-0734 Y",
			expectError:      false,
			expectedEmail:    "Duser1_1@example.fake",
			expectedPhone:    "000-6320-0734",
			expectedCreditUp: true,
		},
		{
			name:        "필드 부족",
			line:        "only_email@example.com", // 전화번호, 신용점수 없음
			expectError: true,
		},
		{
			name:             "신용점수 하락 사용자 정상 파싱",
			line:             "Duser206226_26@example.fake                   000-1815-2005   N",
			expectError:      false,
			expectedEmail:    "Duser206226_26@example.fake",
			expectedPhone:    "000-1815-2005",
			expectedCreditUp: false,
		},
		{
			name:             "앞뒤 공백이 있는 라인",
			line:             "  Duser206226_26@example.fake           000-1815-2005   N  ",
			expectError:      false,
			expectedEmail:    "Duser206226_26@example.fake",
			expectedPhone:    "000-1815-2005",
			expectedCreditUp: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When: 라인 파싱 실행
			user, err := parser.parseLine(tc.line)

			// Then: 결과 검증
			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			assert.Equal(t, tc.expectedEmail, user.Email)
			assert.Equal(t, tc.expectedPhone, user.PhoneNumber)
			assert.Equal(t, tc.expectedCreditUp, user.CreditUp)
		})
	}
}

func TestFileParser_ParseUsers_WithValidFile(t *testing.T) {
	// Given: 테스트 파일 생성
	testData := `Duser780641_29@example.fake          000-0420-2932   Y
Duser206226_26@example.fake                        000-1815-2005   N
Duser468598_84@example.fake                        000-1311-1060   Y`

	tmpFile, err := os.CreateTemp("", "test_data_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())
	ctx := context.Background()

	// When: 파일 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 결과 검증
	require.NoError(t, err)
	assert.Len(t, users, 3)

	// 첫 번째 사용자 검증
	assert.Equal(t, "Duser780641_29@example.fake", users[0].Email)
	assert.True(t, users[0].CreditUp)

	// 두 번째 사용자 검증
	assert.False(t, users[1].CreditUp)
}

func TestFileParser_ParseUsers_WithEmptyLines(t *testing.T) {
	// Given: 빈 라인이 포함된 테스트 파일
	testData := `Duser780641_29@example.fake                        000-0420-2932   Y

Duser206226_26@example.fake                        000-1815-2005   N
   
Duser468598_84@example.fake                        000-1311-1060   Y`

	tmpFile, err := os.CreateTemp("", "test_data_empty_lines_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())
	ctx := context.Background()

	// When: 파일 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 빈 라인이 무시되고 3명의 사용자만 파싱되어야 함
	require.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestFileParser_ParseUsers_WithContext(t *testing.T) {
	// Given: 큰 테스트 파일 생성
	testData := strings.Repeat("Duser780641_29@example.fake                000-0420-2932   Y\n", 1000)

	tmpFile, err := os.CreateTemp("", "test_large_data_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())

	// Given: 즉시 취소되는 컨텍스트
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	// When: 파싱 실행 (즉시 취소 예상)
	users, err := parser.ParseUsers(ctx)

	// Then: 컨텍스트 에러 확인
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, users)
}

func TestFileParser_ParseUsers_FileNotFound(t *testing.T) {
	// Given: 존재하지 않는 파일 경로
	parser := NewFileParser("존재하지않는파일.txt")
	ctx := context.Background()

	// When: 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 에러 확인
	assert.Error(t, err)
	assert.Nil(t, users)
}

func TestFileParser_ParseUsers_InvalidLineFormat(t *testing.T) {
	// Given: 잘못된 형식의 라인이 포함된 파일
	testData := `Duser780641_29@example.fake              000-0420-2932   Y
잘못된_짧은_라인
Duser206226_26@example.fake               000-1815-2005   N`

	tmpFile, err := os.CreateTemp("", "test_invalid_format_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())
	ctx := context.Background()

	// When: 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 에러 확인 (잘못된 라인 때문에 파싱 실패)
	assert.Error(t, err)
	assert.Nil(t, users)
}
