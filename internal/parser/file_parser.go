package parser

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"banksalad-backend-task/internal/domain"
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
		return nil, errors.Wrap(err, "파일을 열 수 없습니다")
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.WithError(err).Error("failed to close file")
		}
	}()

	users := make([]*domain.User, 0, 8000)
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
			return nil, errors.Wrapf(err, "%d번째 라인 파싱 오류", lineNumber)
		}

		users = append(users, user)
	}

	// Then: 스캔 에러 확인
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "파일 읽기 오류")
	}

	return users, nil
}

func (fp *FileParser) parseLine(line string) (*domain.User, error) {
	// 공백을 기준으로 필드 분리
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return nil, errors.New("필드가 부족합니다: 최소 3개 필요")
	}

	email := fields[0]
	phoneNumber := fields[1]
	creditUpStr := fields[2]

	creditUp := creditUpStr == "Y"

	user, err := domain.NewUser(email, phoneNumber, creditUp)
	if err != nil {
		return nil, errors.Wrap(err, "사용자 객체 생성 실패")
	}

	return user, nil
}
