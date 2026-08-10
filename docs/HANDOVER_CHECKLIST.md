# Checklist bàn giao FlashX MVP

## 1. Source code
- [ ] Rider app source.
- [ ] Driver app source.
- [ ] Admin source.
- [ ] Go API source.
- [ ] Database migrations.
- [ ] Infrastructure/deployment config.
- [ ] `.env.example`, không chứa production secret.

## 2. Build/release
- [ ] Rider Android release build.
- [ ] Rider iOS archive/build path documented.
- [ ] Driver Android release build.
- [ ] Driver iOS archive/build path documented.
- [ ] Admin production build/deploy.
- [ ] API production image/build.
- [ ] Store submission package/checklist nếu nằm trong hợp đồng.

## 3. Environments
- [ ] Local.
- [ ] Staging.
- [ ] Production.
- [ ] Database backup policy.
- [ ] Redis production policy.
- [ ] Object storage.
- [ ] DNS/TLS.

## 4. Third-party ownership
Khách hàng phải sở hữu hoặc được bàn giao quyền quản trị:
- [ ] Maps provider project/billing.
- [ ] SMS/OTP provider.
- [ ] Push/Firebase/Apple config.
- [ ] Payment gateway merchant account nếu có.
- [ ] Cloud account.
- [ ] Object storage.
- [ ] Apple Developer.
- [ ] Google Play Console.
- [ ] Monitoring/Error tracking account nếu dùng SaaS.

## 5. Security
- [ ] Production secrets không nằm trong Git.
- [ ] Key rotation procedure.
- [ ] Admin RBAC tested.
- [ ] KYC access restricted/audited.
- [ ] TLS enabled.
- [ ] Backup restore test.
- [ ] Dependency/security scan theo release policy.

## 6. Business acceptance
- [ ] Rider login.
- [ ] Rider estimate.
- [ ] Rider creates trip.
- [ ] Approved Driver online.
- [ ] Driver sends valid location.
- [ ] Dispatch finds candidate.
- [ ] Offer reaches Driver.
- [ ] Exactly one Driver accepts.
- [ ] Rider sees assigned Driver.
- [ ] Driver Arrived -> Start -> Complete.
- [ ] Rider receives completed fare.
- [ ] Trip history persists.
- [ ] Rating works.
- [ ] Admin sees Rider/Driver/Trip.
- [ ] Admin KYC decision works.
- [ ] Admin pricing configuration works.

## 7. Failure acceptance
- [ ] No driver found.
- [ ] Driver rejects/offer expires.
- [ ] Two drivers accept concurrently.
- [ ] Rider cancels while searching.
- [ ] Network disconnect/reconnect.
- [ ] GPS disabled/stale.
- [ ] App background/resume.
- [ ] Duplicate create/complete request.
- [ ] Provider timeout.

## 8. Documentation
- [ ] Product/PRD.
- [ ] Architecture/ADR.
- [ ] Domain/data model.
- [ ] API/OpenAPI.
- [ ] Realtime/dispatch.
- [ ] Maps/third-party cost.
- [ ] UX/UI system.
- [ ] Security/privacy.
- [ ] Infra/DevOps.
- [ ] Testing report.
- [ ] Deployment/rollback runbook.
- [ ] Known limitations.

## 9. Release gate
Bàn giao MVP chỉ được coi là hoàn tất khi core ride flow chạy end-to-end trên staging/production target và không còn blocker/critical bug đã biết.
