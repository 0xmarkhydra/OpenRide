# TestFlight release automation

FlashX has a single release workflow for both iOS applications:

- Rider: `com.flashx.flashxRider`
- Driver: `com.flashx.flashxDriver`

Run **FlashX TestFlight** from the GitHub Actions tab, select Rider, Driver, or both, enter the marketing version and optional release notes. To publish both automatically, push a tag named `testflight-v*` (for example, `testflight-v1.0.0`). Each run uses its GitHub run number as the iOS build number, so uploads remain monotonically increasing.

## One-time GitHub setup

Set the following repository **Actions secrets**. Do not store any of these values in the repository.

| Secret | Value |
| --- | --- |
| `APP_STORE_CONNECT_API_KEY_ID` | Key ID of an App Store Connect API key with App Manager access. |
| `APP_STORE_CONNECT_ISSUER_ID` | Issuer ID shown for that API key. |
| `APP_STORE_CONNECT_API_KEY_BASE64` | Base64 encoding of the downloaded `.p8` private-key file. |
| `API_BASE_URL_STAGING` | HTTPS URL of the staging FlashX API. |
| `MATCH_GIT_URL` | Private Git URL for the encrypted Fastlane Match signing repository. |
| `MATCH_PASSWORD` | Encryption password for that signing repository. |
| `MATCH_GIT_PRIVATE_KEY` | Optional SSH deploy key with read/write access to the signing repository. Required if `MATCH_GIT_URL` uses SSH. |

Create the Match repository once as a private, empty Git repository. The workflow creates and maintains App Store certificates and provisioning profiles there, encrypted with `MATCH_PASSWORD`. Grant the API key access to both apps in App Store Connect and make sure both bundle IDs exist in the Apple Developer account for Team `5SW8LWW4HW`.

## First release check

Before the first upload, create the two app records in App Store Connect with their exact bundle IDs. Run the workflow for one app and confirm that Apple accepts the build. TestFlight build processing can take several minutes; the workflow intentionally finishes after a successful upload rather than waiting for Apple processing.
