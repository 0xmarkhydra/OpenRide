# Public repository audit criteria

OpenRide should be easy for an experienced developer to evaluate without having to reverse-engineer which claims are real and which are roadmap.

The repository is considered ready for a public pre-1.0 review when:

- project status is explicit and separates implemented, compatibility and planned work;
- public branding is OpenRide while legacy identifiers are documented rather than hidden;
- modules can resolve without relying silently on `go.work`;
- CI validates code, packages, Docker images and a bootable microservices slice;
- service boundaries own their data and migrations;
- database integrity is enforced in additive migrations;
- API and event compatibility policies are documented;
- security, contribution, versioning and release processes are visible;
- the license is machine-detectable and complete;
- examples do not imply endpoints or production readiness that do not exist yet.

This document is not a certification of production safety. OpenRide remains pre-1.0 until the operational, legal, safety, privacy, insurance and payment requirements of real passenger transport are validated for a target deployment.
