# Security Policy

OpenRide handles mobility, identity, location and payment-adjacent workflows. Security issues must be treated as product issues, not just code bugs.

## Supported versions

OpenRide is currently pre-1.0. Security fixes are applied to the active `main` branch unless a release explicitly documents otherwise.

## Reporting a vulnerability

Please **do not open a public GitHub issue** for vulnerabilities that could expose users, drivers, credentials, precise location data, payments, authentication, operator access or infrastructure.

Use GitHub's private vulnerability reporting / Security Advisory flow for this repository when available. If private reporting is unavailable, contact the repository maintainers privately before publishing technical details.

A useful report includes:

- affected commit or version;
- affected component;
- reproduction steps;
- realistic impact;
- logs or screenshots with secrets and personal data removed;
- suggested mitigation, if known.

## Security priorities

Issues receive elevated priority when they involve:

- authentication or authorization bypass;
- OTP/session/JWT compromise;
- operator/admin privilege escalation;
- precise location disclosure;
- driver/rider identity or KYC data exposure;
- object-storage authorization or presigned URL abuse;
- payment manipulation or replay;
- marketplace quote/agreement tampering;
- remote code execution, SQL injection or SSRF;
- secrets committed to source control;
- cross-instance data leakage when multi-instance support is enabled.

## Responsible disclosure

Please allow maintainers a reasonable opportunity to reproduce, fix and publish a remediation before public disclosure. We will aim to acknowledge valid reports, communicate material status changes and credit reporters who want attribution.

## Deployment responsibility

OpenRide is self-hostable software. Operators are responsible for secure infrastructure, TLS, secret management, backups, access control, local legal requirements and timely security updates. Development defaults and mock providers must not be used as production security controls.
