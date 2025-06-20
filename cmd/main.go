package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"banksalad-backend-task/internal/domain"
	"banksalad-backend-task/internal/parser"
	"banksalad-backend-task/internal/processor"
	"banksalad-backend-task/internal/service"
)

func main() {
	// 시작 시간 기록
	startTime := time.Now()

	// 컨텍스트 설정 (Ctrl+C로 중단 가능)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 시그널 처리 (우아한 종료)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\n프로그램을 종료합니다...")
		cancel()
	}()

	fmt.Println("=== 뱅크샐러드 신용점수 알림 시스템 ===")
	fmt.Println()

	// 출력 디렉토리 생성
	if err := ensureOutputDirectory(); err != nil {
		log.Fatalf("출력 디렉토리 생성 실패: %v", err)
	}

	// 1단계: 파일 파싱
	fmt.Println("1단계: 데이터 파일 파싱 중...")
	users, err := parseDataFile(ctx)
	if err != nil {
		log.Fatalf("파일 파싱 실패: %v", err)
	}
	fmt.Printf("✓ 총 %d명의 사용자 데이터를 읽었습니다.\n\n", len(users))

	// 2단계: 신용점수 상승 사용자 필터링
	fmt.Println("2단계: 신용점수 상승 사용자 필터링 중...")
	eligibleUsers := filterEligibleUsers(users)
	fmt.Printf("✓ 신용점수 상승 사용자: %d명\n\n", len(eligibleUsers))

	// 3단계: 중복 제거
	fmt.Println("3단계: 중복 사용자 제거 중...")
	uniqueUsers := removeDuplicates(eligibleUsers)
	removedDuplicates := len(eligibleUsers) - len(uniqueUsers)
	if removedDuplicates > 0 {
		fmt.Printf("✓ 중복 제거 후: %d명 (중복 %d명 제거)\n\n", len(uniqueUsers), removedDuplicates)
	} else {
		fmt.Printf("✓ 중복 제거 후: %d명 (중복 없음)\n\n", len(uniqueUsers))
	}

	// 4단계: 알림 전송
	var emailSuccess, smsSuccess int
	if len(uniqueUsers) > 0 {
		fmt.Println("4단계: 알림 전송 중...")
		fmt.Printf("- SMS 속도 제한: 초당 100건\n")
		fmt.Printf("- 이메일: 병렬 전송 (제한 없음)\n")

		emailSuccess, smsSuccess, err = sendNotifications(ctx, uniqueUsers)
		if err != nil {
			log.Fatalf("알림 전송 실패: %v", err)
		}

		// 실제 성공 수 출력
		bothSuccess := min(emailSuccess, smsSuccess)
		fmt.Printf("✓ 알림 전송 완료: 이메일 %d명, SMS %d명, 양쪽 모두 성공 %d명\n\n",
			emailSuccess, smsSuccess, bothSuccess)
	} else {
		fmt.Println("알림을 보낼 사용자가 없습니다.\n")
	}

	// 결과 요약
	printSummary(startTime, len(users), len(eligibleUsers), len(uniqueUsers), emailSuccess, smsSuccess)
}

func ensureOutputDirectory() error {
	if err := os.MkdirAll("files/output", 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	return nil
}

func parseDataFile(ctx context.Context) ([]*domain.User, error) {
	fileParser := parser.NewFileParser("files/input/data.txt")
	users, err := fileParser.ParseUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("데이터 파일 파싱 중 오류: %w", err)
	}
	return users, nil
}

func filterEligibleUsers(users []*domain.User) []*domain.User {
	creditProcessor := processor.NewCreditProcessor()
	return creditProcessor.FilterEligibleUsers(users)
}

func removeDuplicates(users []*domain.User) []*domain.User {
	duplicateFilter := processor.NewDuplicateFilter()
	return duplicateFilter.FilterDuplicates(users)
}

func sendNotifications(ctx context.Context, users []*domain.User) (int, int, error) {
	notificationManager := service.NewNotificationManager()
	defer notificationManager.Close()

	emailSuccess, smsSuccess, err := notificationManager.SendNotifications(ctx, users)
	if err != nil {
		return emailSuccess, smsSuccess, fmt.Errorf("알림 전송 중 오류: %w", err)
	}

	return emailSuccess, smsSuccess, nil
}

func printSummary(startTime time.Time, totalUsers, eligibleUsers, uniqueUsers, emailSuccess, smsSuccess int) {
	duration := time.Since(startTime)

	fmt.Println("=== 실행 결과 요약 ===")
	fmt.Printf("총 처리 시간: %v\n", duration)
	fmt.Printf("전체 사용자: %d명\n", totalUsers)
	fmt.Printf("신용점수 상승: %d명 (%.1f%%)\n",
		eligibleUsers, float64(eligibleUsers)/float64(totalUsers)*100)
	fmt.Printf("중복 제거 후: %d명\n", uniqueUsers)
	fmt.Printf("이메일 전송 성공: %d명\n", emailSuccess)
	fmt.Printf("SMS 전송 성공: %d명\n", smsSuccess)

	bothSuccess := min(emailSuccess, smsSuccess)
	fmt.Printf("양쪽 모두 성공: %d명\n", bothSuccess)

	if uniqueUsers > 0 {
		avgTimePerUser := duration / time.Duration(uniqueUsers)
		fmt.Printf("사용자당 평균 처리 시간: %v\n", avgTimePerUser)
	}

	fmt.Println("\n=== 출력 파일 ===")

	// 파일 존재 여부 확인
	checkOutputFiles()
}

func checkOutputFiles() {
	files := []string{
		"files/output/notified_emails.txt",
		"files/output/notified_phone_numbers.txt",
	}

	for _, filePath := range files {
		if info, err := os.Stat(filePath); err == nil {
			fmt.Printf("✓ %s (크기: %d bytes)\n", filePath, info.Size())
		} else {
			fmt.Printf("✗ %s (파일 없음)\n", filePath)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
