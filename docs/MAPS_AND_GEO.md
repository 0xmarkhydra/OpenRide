# Maps & Geospatial Integration

## 1. Nguyên tắc

FlashX không tự xây bản đồ/routing engine trong MVP. Bản đồ, tìm địa điểm, route và ETA sử dụng provider bên thứ ba như Google Maps Platform, Mapbox hoặc HERE thông qua adapter.

## 2. Chức năng cần từ provider

- Mobile map SDK.
- Place autocomplete/search.
- Geocoding/reverse geocoding.
- Route/directions.
- Distance/duration/ETA.
- Navigation handoff hoặc Navigation SDK nếu phase sau cần embedded navigation.

## 3. Ownership chi phí

Chi phí phát triển tích hợp Maps là chi phí dev. Phí API/SDK theo usage do nhà cung cấp Maps thu là chi phí vận hành của khách hàng/doanh nghiệp và phải có tài khoản billing riêng.

Không hard-code key cá nhân của developer vào app production.

## 4. Abstraction

Backend nên có interface logic tương tự:

```text
MapsProvider
- Geocode(query)
- ReverseGeocode(lat,lng)
- Route(origin,destination,mode)
- Matrix(origins,destinations,mode)
```

Mobile map rendering có adapter/config riêng theo SDK nhưng business logic không phụ thuộc provider-specific response ở nhiều nơi.

## 5. Pricing dependency

Fare estimate không lấy straight-line distance. Phải dùng route distance từ provider hoặc routing engine đã được chấp thuận.

Response route nên được normalize:
- distance_m
- duration_s
- encoded polyline/reference
- provider
- provider_request_id nếu có
- computed_at

## 6. Cache

Có thể cache geocoding/place detail phù hợp với điều khoản provider. Không cache trái license/ToS.

Route/ETA có tính thời gian nên cache TTL ngắn hoặc không cache nếu traffic-aware.

## 7. PostGIS

PostGIS dùng cho dữ liệu spatial nội bộ:
- pickup/destination;
- zone/polygon;
- trip spatial analysis;
- service region.

Redis GEO dùng cho driver hot-location search. Hai hệ thống không thay thế nhau.

## 8. Service areas

Khi triển khai theo thành phố, nên lưu service region/zone bằng polygon PostGIS. Trước khi estimate/create trip, backend xác nhận pickup/destination có thuộc vùng phục vụ theo policy.

## 9. Navigation cho Driver

MVP ưu tiên mở external Google Maps/Apple Maps hoặc provider navigation app để giảm độ phức tạp và billing. Embedded turn-by-turn Navigation SDK là phase nâng cấp riêng.

## 10. API key security

- Key server-side lưu trong secret manager/env, không commit Git.
- Mobile key phải giới hạn bundle ID/package name/signing certificate/API scope.
- Dùng separate project/key cho dev/staging/prod.
- Bật quota/budget alert.

## 11. Cost controls

- Autocomplete debounce.
- Không gọi route liên tục khi user chỉ kéo map.
- Reuse route result trong một booking flow khi còn hợp lệ.
- Theo dõi request count theo API/SKU.
- Budget alert và quota theo environment.

## 12. Fallback

Nếu provider lỗi:
- search/estimate có thể báo tạm thời không khả dụng;
- trip đang chạy vẫn tiếp tục dựa trên state nội bộ;
- không để provider outage làm corrupt trip state.

## 13. Provider switch

Switch provider là project riêng nếu UI/SDK khác nhau. Backend abstraction giảm impact cho routing/geocoding nhưng không đảm bảo thay mobile SDK chỉ bằng config.
