#!/usr/bin/env python3
"""End-to-end smoke test for the three FlashX marketplace demo services.

Runs only against a development/staging API where OTP debug codes and the
/v1/dev/drivers bootstrap endpoint are enabled. It creates isolated demo data
with a unique suffix and intentionally leaves the completed records available
for the Admin Operations screen.
"""

from __future__ import annotations

import json
import os
import sys
import time
import urllib.error
import urllib.request
from typing import Any

BASE_URL = os.environ.get("API_BASE_URL", "http://127.0.0.1:8080").rstrip("/")
PICKUP = {"lat": 19.8060, "lng": 105.7760}
DESTINATION = {"lat": 19.8150, "lng": 105.7890}


class SmokeFailure(RuntimeError):
    pass


def call(method: str, path: str, *, body: dict[str, Any] | None = None,
         headers: dict[str, str] | None = None,
         expected: tuple[int, ...] = (200,)) -> tuple[int, Any]:
    payload = None if body is None else json.dumps(body).encode("utf-8")
    request_headers = {"Accept": "application/json"}
    if body is not None:
        request_headers["Content-Type"] = "application/json"
    if headers:
        request_headers.update(headers)
    req = urllib.request.Request(
        BASE_URL + path,
        data=payload,
        method=method,
        headers=request_headers,
    )
    try:
        with urllib.request.urlopen(req, timeout=8) as response:
            status = response.status
            raw = response.read()
    except urllib.error.HTTPError as exc:
        status = exc.code
        raw = exc.read()
    except OSError as exc:
        raise SmokeFailure(f"Cannot connect to {BASE_URL}: {exc}") from exc

    parsed: Any = None
    if raw:
        try:
            parsed = json.loads(raw.decode("utf-8"))
        except json.JSONDecodeError:
            parsed = raw.decode("utf-8", errors="replace")
    if status not in expected:
        raise SmokeFailure(
            f"{method} {path} -> HTTP {status}, expected {expected}: {parsed}"
        )
    return status, parsed


def data(payload: Any) -> Any:
    if not isinstance(payload, dict) or "data" not in payload:
        raise SmokeFailure(f"Response has no data envelope: {payload}")
    return payload["data"]


def login_rider(phone: str) -> tuple[str, str]:
    _, requested = call(
        "POST", "/v1/auth/otp/request",
        body={"phone": phone, "role": "rider"},
        expected=(202,),
    )
    request_data = data(requested)
    code = request_data.get("debug_code")
    challenge_id = request_data.get("challenge_id")
    if not code or not challenge_id:
        raise SmokeFailure(
            "Development OTP debug code is unavailable. Run against a demo/dev API."
        )
    _, verified = call(
        "POST", "/v1/auth/otp/verify",
        body={"challenge_id": challenge_id, "role": "rider", "code": code},
    )
    auth = data(verified)
    token = auth.get("tokens", {}).get("access_token")
    rider_id = auth.get("actor", {}).get("id")
    if not token or not rider_id:
        raise SmokeFailure(f"Invalid rider auth response: {auth}")
    return token, rider_id


def create_vehicle(token: str, *, vehicle_type: str, plate: str) -> dict[str, Any]:
    is_car = vehicle_type == "car"
    _, response = call(
        "POST", "/v1/rider/vehicles",
        headers={"Authorization": f"Bearer {token}"},
        body={
            "type": vehicle_type,
            "license_plate": plate,
            "brand": "Toyota" if is_car else "Honda",
            "model": "Vios" if is_car else "Vision",
            "year": 2024,
            "color": "Trắng" if is_car else "Đen",
            "transmission": "automatic" if is_car else "n/a",
            "seats": 5 if is_car else 2,
            "notes": "Dữ liệu smoke test demo Bộ Công Thương",
            "photo_object_key": "",
        },
        expected=(201,),
    )
    return data(response)


def create_driver(driver_id: str, service_type: str) -> dict[str, str]:
    _, response = call(
        "POST", "/v1/dev/drivers",
        body={
            "id": driver_id,
            "full_name": f"Tài xế demo {service_type}",
            "service_type": service_type,
        },
        expected=(201,),
    )
    driver = data(response)
    headers = {"X-Dev-Driver-ID": driver["id"]}
    call(
        "POST", "/v1/driver/availability",
        headers=headers,
        body={"status": "online"},
    )
    call(
        "POST", "/v1/driver/location",
        headers=headers,
        body={
            "lat": PICKUP["lat"] + 0.0003,
            "lng": PICKUP["lng"] + 0.0003,
            "accuracy_m": 5,
        },
    )
    return headers


def create_job(token: str, suffix: str, service_type: str, vehicle_id: str) -> dict[str, Any]:
    _, estimate = call(
        "POST", "/v1/trips/estimate",
        headers={"Authorization": f"Bearer {token}"},
        body={
            "pickup": PICKUP,
            "destination": DESTINATION,
            "service_type": service_type,
            "customer_vehicle_id": vehicle_id,
            "booking_mode": "immediate",
        },
    )
    estimate_data = data(estimate)
    if int(estimate_data.get("fare", {}).get("total_minor", 0)) <= 0:
        raise SmokeFailure(f"Invalid fare estimate for {service_type}: {estimate_data}")

    _, created = call(
        "POST", "/v1/trips",
        headers={
            "Authorization": f"Bearer {token}",
            "Idempotency-Key": f"smoke-{service_type}-{suffix}",
        },
        body={
            "pickup": PICKUP,
            "destination": DESTINATION,
            "service_type": service_type,
            "customer_vehicle_id": vehicle_id,
            "booking_mode": "immediate",
        },
        expected=(201,),
    )
    job = data(created)
    if job.get("status") != "searching":
        raise SmokeFailure(f"New {service_type} job is not searching: {job}")
    return job


def current_offer(driver_headers: dict[str, str]) -> dict[str, Any]:
    for _ in range(10):
        status, response = call(
            "GET", "/v1/driver/offers/current",
            headers=driver_headers,
            expected=(200, 204),
        )
        if status == 200:
            return data(response)
        time.sleep(0.15)
    raise SmokeFailure("Driver did not receive an offer")


def driver_action(driver_headers: dict[str, str], job_id: str, action: str,
                  body: dict[str, Any] | None = None) -> dict[str, Any]:
    _, response = call(
        "POST", f"/v1/driver/trips/{job_id}/{action}",
        headers=driver_headers,
        body={} if body is None else body,
    )
    return data(response)


def run_service(token: str, suffix: str, service_type: str, vehicle_id: str,
                driver_headers: dict[str, str]) -> dict[str, Any]:
    job = create_job(token, suffix, service_type, vehicle_id)
    job_id = job["id"]
    offer_payload = current_offer(driver_headers)
    offer = offer_payload.get("offer", {})
    offered_job = offer_payload.get("trip", {})
    offered_vehicle = offer_payload.get("vehicle", {})
    if offered_job.get("id") != job_id:
        raise SmokeFailure(
            f"Driver received wrong job: expected {job_id}, got {offered_job.get('id')}"
        )
    if offered_vehicle.get("id") != vehicle_id:
        raise SmokeFailure(
            f"Offer does not include the selected customer vehicle for {service_type}"
        )

    _, accepted = call(
        "POST", f"/v1/driver/offers/{offer['id']}/accept",
        headers=driver_headers,
        body={},
    )
    accepted_job = data(accepted).get("trip", {})
    if accepted_job.get("status") != "accepted":
        raise SmokeFailure(f"Offer accept failed for {service_type}: {accepted_job}")

    if service_type == "vehicle_inspection_assist":
        steps: list[tuple[str, dict[str, Any] | None]] = [
            ("arriving", None),
            ("arrived", None),
            ("vehicle-received", None),
            ("start", None),
            ("arrive-inspection", None),
            ("start-inspection", None),
            ("complete-inspection", {"result": "passed"}),
            ("returning", None),
            ("arrived-return", None),
            ("handover", None),
            ("complete", None),
        ]
    else:
        steps = [
            ("arriving", None),
            ("arrived", None),
            ("vehicle-received", None),
            ("start", None),
            ("handover", None),
            ("complete", None),
        ]

    latest: dict[str, Any] = accepted_job
    for action, body in steps:
        latest = driver_action(driver_headers, job_id, action, body)

    if latest.get("status") != "completed":
        raise SmokeFailure(f"{service_type} did not complete: {latest}")

    _, snapshot = call(
        "GET", f"/v1/trips/{job_id}",
        headers={"Authorization": f"Bearer {token}"},
    )
    final_job = data(snapshot)
    if final_job.get("status") != "completed":
        raise SmokeFailure(f"Rider snapshot is not completed: {final_job}")
    if service_type == "vehicle_inspection_assist" and final_job.get("inspection_result") != "passed":
        raise SmokeFailure(f"Inspection result was not persisted: {final_job}")
    return final_job


def main() -> int:
    suffix = str(int(time.time()))[-8:]
    phone = "09" + suffix
    print(f"FlashX demo smoke against {BASE_URL}")

    _, health = call("GET", "/health")
    _, ready = call("GET", "/ready")
    print(f"  API: {data(health)['status']} / {data(ready)['status']}")

    token, rider_id = login_rider(phone)
    print(f"  Rider: {rider_id}")

    car = create_vehicle(token, vehicle_type="car", plate=f"36A{suffix[-5:]}")
    bike = create_vehicle(token, vehicle_type="motorbike", plate=f"36B{suffix[-5:]}")
    print(f"  Vehicles: {car['license_plate']} + {bike['license_plate']}")

    services = [
        ("designated_driver_car", car["id"]),
        ("designated_driver_bike", bike["id"]),
        ("vehicle_inspection_assist", car["id"]),
    ]

    results: list[dict[str, Any]] = []
    for index, (service_type, vehicle_id) in enumerate(services, start=1):
        driver_id = f"smoke-{index}-{suffix}"
        driver_headers = create_driver(driver_id, service_type)
        final_job = run_service(
            token, f"{suffix}-{index}", service_type, vehicle_id, driver_headers
        )
        results.append(final_job)
        fare = final_job.get("final_fare_minor") or final_job.get("estimated_fare_minor")
        print(f"  PASS {service_type}: {final_job['id']} / {fare} VND")

    _, history = call(
        "GET", "/v1/trips?limit=20",
        headers={"Authorization": f"Bearer {token}"},
    )
    completed_ids = {item.get("id") for item in data(history) if item.get("status") == "completed"}
    expected_ids = {item["id"] for item in results}
    if not expected_ids.issubset(completed_ids):
        raise SmokeFailure("Completed demo jobs are missing from rider history")

    print("\nALL 3 FLASHX DEMO SERVICES PASSED")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except SmokeFailure as exc:
        print(f"\nSMOKE FAILED: {exc}", file=sys.stderr)
        raise SystemExit(1)
