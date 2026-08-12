# Object Storage & Direct Media Upload

## 1. Quyết định kiến trúc

FlashX **không proxy file media qua backend**.

Luồng bắt buộc:

```text
Flutter / Admin Client
    |
    | 1. Xin presigned request (JSON metadata only)
    v
FlashX Go API
    |
    | 2. Ký request bằng S3 credential phía server
    v
Client nhận presigned PUT/GET
    |
    | 3. Upload / download trực tiếp
    v
S3-compatible Object Storage
    |
    | 4. Client báo metadata upload hoàn tất
    v
FlashX Go API -> PostgreSQL
```

Backend không nhận multipart/file binary, không giữ file trong RAM và không stream file sang object storage.

## 2. Provider

Adapter dùng S3-compatible API. Có thể dùng VNDATA Object Storage hoặc provider S3-compatible khác bằng cách đổi:

- `S3_ENDPOINT`
- `S3_REGION`
- `S3_BUCKET`
- `S3_ACCESS_KEY_ID`
- `S3_SECRET_ACCESS_KEY`
- `S3_FORCE_PATH_STYLE`

Bucket phải private.

## 3. Configuration

```env
OBJECT_STORAGE_PROVIDER=s3
S3_ENDPOINT=https://s3-hcm-r2.s3cloud.vn
S3_REGION=us-east-1
S3_BUCKET=flashx-media
S3_ACCESS_KEY_ID=***
S3_SECRET_ACCESS_KEY=***
S3_FORCE_PATH_STYLE=true
S3_PRESIGN_TTL_SECONDS=600
```

Không commit credential thật.

Production không được chạy với `OBJECT_STORAGE_PROVIDER=disabled`.

## 4. Driver KYC flow

### Step 1 — xin chữ ký upload

```http
POST /v1/driver/documents/upload-url
Authorization: Bearer <driver-access-token>
Content-Type: application/json

{
  "document_type": "driver_license",
  "filename": "gplx.jpg",
  "content_type": "image/jpeg",
  "size_bytes": 1842310
}
```

Response:

```json
{
  "data": {
    "document_id": "doc_...",
    "document_type": "driver_license",
    "object_key": "flashx/kyc/drivers/drv_.../driver_license/2026/08/doc_...-gplx.jpg",
    "upload": {
      "method": "PUT",
      "url": "https://...signed...",
      "headers": {
        "Content-Type": ["image/jpeg"]
      },
      "expires_at": "2026-08-11T01:20:00Z"
    }
  }
}
```

### Step 2 — client upload trực tiếp

Client gọi URL đã ký:

```http
PUT <upload.url>
Content-Type: image/jpeg
Content-Length: 1842310

<raw file bytes>
```

Request này đi thẳng từ thiết bị tới object storage. FlashX API không nằm trên đường truyền file.

### Step 3 — lưu metadata sau upload

```http
POST /v1/driver/documents/complete
Authorization: Bearer <driver-access-token>
Content-Type: application/json

{
  "document_id": "doc_...",
  "document_type": "driver_license",
  "object_key": "flashx/kyc/drivers/drv_.../driver_license/2026/08/doc_...-gplx.jpg",
  "filename": "gplx.jpg",
  "content_type": "image/jpeg",
  "size_bytes": 1842310
}
```

Backend chỉ lưu metadata vào PostgreSQL. Backend không tải lại object để proxy hoặc lưu bản sao.

## 5. Xem file

Driver hoặc Admin không nhận public URL vĩnh viễn.

Driver:

```http
GET /v1/driver/documents/{documentID}/view-url
```

Admin:

```http
GET /v1/admin/drivers/{driverID}/documents
```

Backend kiểm tra authorization rồi chỉ cấp presigned GET URL ngắn hạn.

## 6. KYC document types MVP

- `identity_front`
- `identity_back`
- `driver_license`
- `vehicle_registration`
- `vehicle_insurance`
- `portrait`

Content type cho phép:

- `image/jpeg`
- `image/png`
- `image/webp`
- `application/pdf`

Mỗi file KYC tối đa 15 MB ở bước xin chữ ký.

## 7. Object key namespace

Không cho client tự chọn key tùy ý.

Backend sinh key:

```text
flashx/kyc/drivers/{driver_id}/{document_type}/{YYYY}/{MM}/{document_id}-{safe_filename}
```

Lợi ích:

- chống ghi đè object ngoài namespace của tài xế;
- dễ lifecycle/retention theo prefix;
- dễ audit;
- không dùng filename do client gửi làm authority.

## 8. CORS của bucket

Với Flutter native thường không bị browser CORS như web. Nếu Admin/web upload trực tiếp, object storage phải cho phép origin production của FlashX với các method cần thiết như `PUT`, `GET`, và các signed headers tương ứng.

Không cấu hình `AllowedOrigin=*` cho production nếu không cần thiết.

## 9. Security rules

- S3 access key/secret chỉ tồn tại ở backend/secret manager.
- Mobile/Admin browser không bao giờ nhận S3 secret.
- Client chỉ nhận URL tạm thời có scope cho đúng object key/method.
- Bucket private, không public-read.
- Signed URL TTL ngắn; mặc định MVP 10 phút.
- Backend kiểm tra actor trước khi ký GET.
- Metadata KYC không public.
- Không log presigned URL đầy đủ vì URL chứa chữ ký tạm thời.
- Không log raw document.

## 10. Upload lớn

KYC MVP dùng presigned single PUT vì file nhỏ.

Video/audio/file lớn trong tương lai phải dùng multipart direct upload:

```text
Client -> API create multipart/sign parts
Client -> S3 UploadPart trực tiếp
Client -> API request/sign completion workflow
```

Vẫn giữ nguyên nguyên tắc: file bytes không đi qua FlashX backend.

## 11. Database

`driver_documents` lưu metadata:

- document id;
- driver id;
- document type;
- object key;
- filename;
- content type;
- size;
- review status/note;
- created/updated/reviewed timestamps.

Object storage giữ bytes; PostgreSQL giữ business metadata.

## 12. Nguyên tắc bắt buộc

**FlashX Backend = control plane / signing plane.**

**S3-compatible Object Storage = data plane cho media.**

Không thêm endpoint kiểu `POST /upload` nhận multipart vào Go API trừ khi có quyết định kiến trúc mới được ghi thành ADR.
