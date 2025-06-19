package parser

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestFileParser_ParseLine(t *testing.T) {
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
			line:             "Duser780641_29@example.fake                        000-0420-2932   Y",
			expectError:      false,
			expectedEmail:    "Duser780641_29@example.fake",
			expectedPhone:    "000-0420-2932",
			expectedCreditUp: true,
		},
		{
			name:             "신용점수 하락 사용자 정상 파싱",
			line:             "Duser206226_26@example.fake                        000-1815-2005   N",
			expectError:      false,
			expectedEmail:    "Duser206226_26@example.fake",
			expectedPhone:    "000-1815-2005",
			expectedCreditUp: false,
		},
		{
			name:        "라인 길이 부족",
			line:        "short",
			expectError: true,
		},
		{
			name:             "앞뒤 공백이 있는 라인",
			line:             "  Duser206226_26@example.fake                      000-1815-2005   N  ",
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
				if err == nil {
					t.Error("에러가 예상되었지만 발생하지 않음")
				}
				return
			}

			if err != nil {
				t.Errorf("예상치 못한 에러: %v", err)
				return
			}

			if user.Email != tc.expectedEmail {
				t.Errorf("이메일 불일치: 예상=%s, 실제=%s", tc.expectedEmail, user.Email)
			}

			if user.PhoneNumber != tc.expectedPhone {
				t.Errorf("전화번호 불일치: 예상=%s, 실제=%s", tc.expectedPhone, user.PhoneNumber)
			}

			if user.CreditUp != tc.expectedCreditUp {
				t.Errorf("신용점수 상승 여부 불일치: 예상=%v, 실제=%v", tc.expectedCreditUp, user.CreditUp)
			}
		})
	}
}

func TestFileParser_ParseUsers_WithValidFile(t *testing.T) {
	// Given: 테스트 파일 생성
	testData := `Duser780641_29@example.fake                        000-0420-2932   Y
Duser206226_26@example.fake                        000-1815-2005   N
Duser468598_84@example.fake                        000-1311-1060   Y`

	tmpFile, err := os.CreateTemp("", "test_data_*.txt")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("테스트 데이터 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())
	ctx := context.Background()

	// When: 파일 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 결과 검증
	if err != nil {
		t.Errorf("예상치 못한 에러: %v", err)
		return
	}

	expectedCount := 3
	if len(users) != expectedCount {
		t.Errorf("사용자 수 불일치: 예상=%d, 실제=%d", expectedCount, len(users))
	}

	// 첫 번째 사용자 검증
	if users[0].Email != "Duser780641_29@example.fake" {
		t.Errorf("첫 번째 사용자 이메일 불일치: 예상=Duser780641_29@example.fake, 실제=%s", users[0].Email)
	}

	if !users[0].CreditUp {
		t.Error("첫 번째 사용자는 신용점수가 상승해야 함")
	}

	// 두 번째 사용자 검증
	if users[1].CreditUp {
		t.Error("두 번째 사용자는 신용점수가 상승하지 않아야 함")
	}
}

func TestFileParser_ParseUsers_WithEmptyLines(t *testing.T) {
	// Given: 빈 라인이 포함된 테스트 파일
	testData := `Duser780641_29@example.fake                        000-0420-2932   Y

Duser206226_26@example.fake                        000-1815-2005   N
   
Duser468598_84@example.fake                        000-1311-1060   Y`

	tmpFile, err := os.CreateTemp("", "test_data_empty_lines_*.txt")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("테스트 데이터 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())
	ctx := context.Background()

	// When: 파일 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 빈 라인이 무시되고 3명의 사용자만 파싱되어야 함
	if err != nil {
		t.Errorf("예상치 못한 에러: %v", err)
		return
	}

	expectedCount := 3
	if len(users) != expectedCount {
		t.Errorf("사용자 수 불일치: 예상=%d, 실제=%d", expectedCount, len(users))
	}
}

func TestFileParser_ParseUsers_WithContext(t *testing.T) {
	// Given: 큰 테스트 파일 생성
	testData := strings.Repeat("Duser780641_29@example.fake                        000-0420-2932   Y\n", 1000)

	tmpFile, err := os.CreateTemp("", "test_large_data_*.txt")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("테스트 데이터 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())

	// Given: 즉시 취소되는 컨텍스트
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	// When: 파싱 실행 (즉시 취소 예상)
	_, err = parser.ParseUsers(ctx)

	// Then: 컨텍스트 에러 확인
	if err == nil {
		t.Error("컨텍스트 취소 에러가 예상되었지만 발생하지 않음")
	}

	if err != context.Canceled {
		t.Errorf("컨텍스트 취소 에러 예상: 예상=context.Canceled, 실제=%v", err)
	}
}

func TestFileParser_ParseUsers_FileNotFound(t *testing.T) {
	// Given: 존재하지 않는 파일 경로
	parser := NewFileParser("존재하지않는파일.txt")
	ctx := context.Background()

	// When: 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 에러 확인
	if err == nil {
		t.Error("존재하지 않는 파일에 대한 에러가 예상되었지만 발생하지 않음")
	}

	if users != nil {
		t.Error("존재하지 않는 파일에 대해 nil 사용자 목록이 예상됨")
	}
}

func TestFileParser_ParseUsers_InvalidLineFormat(t *testing.T) {
	// Given: 잘못된 형식의 라인이 포함된 파일
	testData := `Duser780641_29@example.fake                        000-0420-2932   Y
잘못된_짧은_라인
Duser206226_26@example.fake                        000-1815-2005   N`

	tmpFile, err := os.CreateTemp("", "test_invalid_format_*.txt")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("테스트 데이터 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name())
	ctx := context.Background()

	// When: 파싱 실행
	users, err := parser.ParseUsers(ctx)

	// Then: 에러 확인 (잘못된 라인 때문에 파싱 실패)
	if err == nil {
		t.Error("잘못된 라인 형식에 대한 에러가 예상되었지만 발생하지 않음")
	}

	if users != nil {
		t.Error("잘못된 파일 형식에 대해 nil 사용자 목록이 예상됨")
	}
}
