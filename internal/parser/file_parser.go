package parser

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"banksalad-backend-task/internal/domain"
)

const (
	emailFieldWidth  = 50
	phoneFieldWidth  = 15
	creditFieldWidth = 1
)

type FileParser struct {
	filePath string
}

func NewFileParser(filePath string) *FileParser {
	return &FileParser{
		filePath: filePath,
	}
}

func (fp *FileParser) ParseUsers(ctx context.Context) ([]*domain.User, error) {
	// Given: 파일 열기
	file, err := os.Open(fp.filePath)
	if err != nil {
		return nil, fmt.Errorf("파일을 열 수 없습니다 %s: %w", fp.filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// 파일 닫기 에러는 로그만 남기고 반환하지 않음
		}
	}()

	var users []*domain.User
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	// When: 라인별로 파싱 실행
	for scanner.Scan() {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		lineNumber++
		line := scanner.Text()

		// 빈 라인 스킵
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		user, err := fp.parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("%d번째 라인 파싱 오류: %w", lineNumber, err)
		}

		users = append(users, user)
	}

	// Then: 스캔 에러 확인
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("파일 읽기 오류: %w", err)
	}

	return users, nil
}

func (fp *FileParser) parseLine(line string) (*domain.User, error) {
	expectedLength := emailFieldWidth + phoneFieldWidth + creditFieldWidth
	if len(line) < expectedLength {
		return nil, fmt.Errorf("라인 길이가 부족합니다: 최소 %d자 필요, 실제 %d자",
			expectedLength, len(line))
	}

	// Given: 필드 추출
	email := strings.TrimSpace(line[:emailFieldWidth])
	phoneNumber := strings.TrimSpace(line[emailFieldWidth : emailFieldWidth+phoneFieldWidth])
	creditUpStr := strings.TrimSpace(string(line[len(line)-1])) // 마지막 문자 추출

	// When: 신용점수 상승 여부 판단
	creditUp := creditUpStr == "Y"

	// Then: User 객체 생성 및 반환
	return domain.NewUser(email, phoneNumber, creditUp)
}
