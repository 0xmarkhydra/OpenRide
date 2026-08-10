# Domain Model

## 1. Core domains

FlashX chia domain theo business capability, không chia theo database table.

## 2. Rider

### Entity
`Rider`

Thuộc tính chính:
- id
- phone
- full_name
- status
- created_at

Các trạng thái account tối thiểu:
- active
- suspended
- deleted/disabled theo policy

## 3. Driver

### Aggregate
`Driver`

Bao gồm:
- driver profile
- approval/KYC status
- availability status
- vehicle references

### Approval status
```text
pending -> approved
pending -> rejected
pending -> more_info_required
more_info_required -> pending/approved/rejected
```

### Availability status
```text
offline -> online -> busy -> online -> offline
```

Chỉ driver `approved` mới được `online`.

## 4. Vehicle

Thuộc tính chính:
- id
- driver_id
- service_type
- plate_number
- brand
- model
- active status (sẽ bổ sung khi implement)

Một driver có thể có nhiều vehicle trong tương lai, nhưng MVP nên có một active vehicle tại một thời điểm.

## 5. Trip aggregate

`Trip` là aggregate quan trọng nhất.

Thuộc tính chính hiện tại:
- id
- rider_id
- driver_id nullable trước match
- service_type
- pickup
- destination
- estimated_distance_m
- estimated_duration_s
- estimated_fare_minor
- final_fare_minor
- currency
- status
- timestamps

### State machine

```text
SEARCHING
  | accept
  v
ACCEPTED
  |
  v
ARRIVING
  |
  v
ARRIVED
  |
  v
IN_PROGRESS
  |
  v
COMPLETED

CANCELLED là terminal state theo policy từ các trạng thái cho phép.
```

### Invariants

- Trip `COMPLETED` phải có driver.
- Một trip chỉ có một assigned driver tại một thời điểm.
- Một driver không được có hai active trips trừ khi business model sau này cho phép.
- `final_fare` chỉ được chốt bởi backend.
- State transition trái thứ tự phải bị từ chối.
- Mọi transition quan trọng ghi `trip_status_history`.

## 6. Dispatch

Dispatch không phải chỉ là query nearest driver. Domain này quản lý:
- candidate set;
- candidate ranking;
- trip offer;
- offer expiry;
- accept race;
- assignment lock;
- retry/search expansion.

Entity/record dự kiến:
- DispatchAttempt
- DriverOffer
- Assignment

Các record này có thể ban đầu tồn tại ở Redis + logs; khi cần audit sâu sẽ bổ sung persistent table.

## 7. Location

Location chia thành hai loại:

### LatestDriverLocation
Hot state, Redis:
- driver_id
- lat/lng
- accuracy
- heading
- speed
- captured_at
- received_at

### TripPath
Durable/analytics state, không nhất thiết ghi mọi point. Có thể sample hoặc encode polyline sau này.

## 8. Pricing

Domain pricing nhận:
- service type;
- route distance;
- estimated duration;
- zone/time rules;
- promotions;

Và trả:
- base fare;
- components;
- discount;
- estimated total.

Pricing engine phải deterministic với cùng version/config để dễ audit.

## 9. Payment

Payment state tách khỏi Trip state.

Ví dụ:
```text
pending -> authorized -> captured
pending/authorized -> failed
captured -> refunded (nếu provider hỗ trợ)
```

Trip có thể completed nhưng payment vẫn pending/failed; không ép hai state machine thành một.

## 10. Notification

Notification là orchestration domain, không phải source of truth cho Trip. Nếu push thất bại, trip state vẫn đúng và app phải sync lại từ backend.

## 11. Admin/Audit

Các action nhạy cảm phải audit:
- approve/reject driver;
- suspend rider/driver;
- manual trip intervention;
- pricing change;
- promotion change;
- role/permission change;
- refund/payment override khi có.

Audit record cần actor, action, target, timestamp và metadata tối thiểu.
