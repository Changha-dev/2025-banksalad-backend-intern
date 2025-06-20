## 2025년도 뱅크샐러드 여름 인턴십 백엔드 사전과제

### 뱅크샐러드 신용점수 알림 시스템

신용점수가 상승한 사용자에게 이메일과 SMS 알림을 전송하는 시스템입니다.

## 실행 방법

1. 프로젝트 설정  
``` go mod tidy ```

2. 프로그램 실행  
``` go run cmd/main.go ```

3. 결과 확인  
``` cat files/output/notified_emails.txt ```  
``` cat files/output/notified_phone_numbers.txt ```

## 프로젝트 구조
```
├── cmd/                        # 메인 애플리케이션
│   └── main.go
├── clients/                   # 외부 서비스 클라이언트
│   ├── email_client.go
│   └── sms_client.go
├── internal/
│   ├── domain/                # 사용자 도메인 모델
│   │   └── user.go
│   ├── parser/                # 데이터 파일 파싱
│   │   └── file_parser.go
│   ├── processor/             # 비즈니스 로직
│   │   ├── credit_processor.go   # 신용점수 필터링
│   │   └── duplicate_filter.go   # 중복 제거
│   └── service/               # 서비스 레이어
│       ├── email_service.go
│       ├── sms_service.go
│       ├── notification_manager.go
│       └── rate_limiter.go
├── files/
│   ├── input/
│   │   └── data.txt           # 입력 데이터
│   └── output/                # 출력 결과
│       ├── notified_emails.txt
│       └── notified_phone_numbers.txt
```
