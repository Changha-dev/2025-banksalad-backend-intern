## 2025년도 뱅크샐러드 여름 인턴십 백엔드 사전과제

### 뱅크샐러드 신용점수 알림 시스템

신용점수가 상승한 사용자에게 이메일과 SMS 알림을 전송하는 시스템입니다.

## 실행 방법

``` go run cmd/main.go ```

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
## 구현 방법
### 처리 흐름
`파일 파싱`: 공백으로 구분된 텍스트 파일 읽기 (이메일 + 전화번호 + 상태)

`필터링`: 신용점수 상승(Y) 사용자만 추출

`중복 제거`: 이메일 기준 중복 사용자 제거 (map[string]struct{} 활용)

`알림 전송`: 이메일(병렬) + SMS(속도제한) 동시 전송

### 아키텍처 설계
```
main.go
  ↓
parser (파일 읽기) → domain.User 생성
  ↓
processor (필터링 + 중복제거) → domain.User 처리
  ↓  
service (알림 전송) → clients (email/sms)
  ↓
domain (모든 레이어에서 공통 사용)
```
### 핵심 구현 사항

#### 속도 제한 처리 (SMS)
- **제한**: 초당 100건
- **구현**: Token Bucket 알고리즘 사용
- **동작**: `time.Ticker`로 토큰 보충, 채널을 통한 토큰 관리
- **장점**: 정확한 속도 제어, 컨텍스트 취소 지원

#### 중복 처리 방법
- **기준**: 이메일 주소 기준 중복 제거
- **전략**: `DuplicateStrategy` enum으로 확장 가능 (ByEmail, ByPhone, ByBoth)
- **구현**: `map[string]struct{}`를 활용한 O(1) 중복 검사
- **메모리 효율**: 빈 struct 사용으로 메모리 최적화

#### 뱅크샐러드 개발 컨벤션 적용
- 코드 품질 및 구조
  - Error Handling : `pkg/errors` 패키지를 활용한 Error Stacking 적용
  - Logging : `logrus` 패키지 도입으로 구조화된 로깅
  - Panic Recovery : 고루틴에서의 안전한 패닉 복구 처리
- Import문 정렬 규칙
  1. `Standard library`
  2. `Third-party library`
  3. `Internal library`
- 메모리 최적화
  - 슬라이스 사전 할당 : 예상 크기로 capacity 설정(data.txt 크기인 8000으로 할당)
  - 효율적인 자료구조 : `map[string]struct{}` 활용
- 테스트 전략
  - 테이블 기반 테스트
  - `Given/When/Then` 패턴
  - `testify` 패키지 : assertion과 require를 적극 활용
  - 함수 네이밍 : private(소문자), public(대문자) 구분
- 시간대 처리
  - KST 타임존 : 초기화 단계에서 미리 로딩한 타임존을 활용
- 리소스 관리
  - defer 패턴: 파일 닫기, 리소스 정리에서 에러 로깅

## 이 외 알아두어야 할 사항
- 중복 처리 전략
  - 기본값: ByEmail 설정 (이메일 기준으로 유니크 키 설정)
  - 비즈니스 요구사항에 따라 중복 기준을 유연하게 변경할 수 있도록 설계
    - **ByEmail**: 이메일 기준
    - **ByPhone**: 전화번호 기준
    - **ByBoth**: 이메일 + 전화번호 조합 기준
- 파일 파싱 방식
  - 방법: `strings.Fields()`를 사용한 공백 기준 파싱
  - 장점: 가변 길이 필드 처리에 유리하며, 공백이 많아도 문제없이 처리

