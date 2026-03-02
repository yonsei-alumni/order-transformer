# 기술 노트

## 기술 스택

| 항목 | 내용 |
|---|---|
| 언어 | Go 1.24+ |
| GUI | Fyne v2 (`fyne.io/fyne/v2`) |
| 엑셀 처리 | excelize v2 (`github.com/xuri/excelize/v2`) |
| 빌드 | Makefile (Linux/Windows 크로스 컴파일) |

## 빌드 환경

Fyne은 CGO가 필요합니다.

```bash
# 소스에서 실행
go run .

# 테스트
go test ./...

# Linux 빌드
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o dist/order-transformer-linux-amd64 .

# Windows 빌드 (크로스 컴파일)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o dist/order-transformer-windows-amd64.exe .
```

## 소스 구조

```
order-transformer/
├── main.go              # Fyne GUI (변환/검증 탭)
├── order.go             # 데이터 타입 및 엑셀 읽기
├── transform.go         # 변환 로직
├── validate.go          # 검증 로직
├── excel.go             # 엑셀 쓰기 및 유틸리티
├── transform_test.go    # 변환 관련 테스트
├── validate_test.go     # 검증 관련 테스트
├── form/                # 엑셀 양식 템플릿
│   ├── source-order-form.xlsx    # 구매자 주문 양식
│   └── target-order-form.xlsx    # 출고 지시 양식
├── testdata/            # 테스트 데이터
│   └── source-order-example.xlsx # 예시 주문 데이터
└── Makefile             # 빌드 자동화
```

## 주요 설계

### 데이터 흐름

```
구매자 주문 엑셀 → ReadSourceOrders() → []SourceOrder
                                            ↓
                                      Transform()
                                            ↓
                                      []ShippingOrder → WriteShippingOrders() → 출고 지시 엑셀
```

### 검증 흐름

```
구매자 주문 엑셀 → ReadSourceOrders() → []SourceOrder ─┐
                                                        ├→ Validate() → ValidationResult
출고 지시 엑셀 → ReadShippingOrders() → []ShippingOrder ┘
```

### 엑셀 컬럼 매핑

구매자 주문의 컬럼 인덱스는 `order.go`의 `srcCol*` 상수로 정의되어 있습니다.
변환 규칙의 상세 내용은 `TRANSFORMATION_RULES.md`를 참고하세요.
