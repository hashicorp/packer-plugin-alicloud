# Latest Release

Please refer to [releases](https://github.com/hashicorp/packer-plugin-alicloud/releases) for the latest CHANGELOG information.

---
## 1.2.0 (September 22, 2026)

### New Features

* Added support for HTTPS endpoints in the Alicloud access configuration [GH-150].
* Added support for customer-managed KMS keys when copying images using
  `kms_key_id` and `image_copy_kms_ids` [GH-156].
* Added support for encrypting system and data disks using `disk_kms_key_id` [GH-166].
* Switched ECS instance creation to the `RunInstances` API [GH-166].
* Added KMS key configuration support when launching ECS instances [GH-166].

### Maintenance

* Updated the Packer Plugin SDK from v0.5.4 to v0.6.0.
* Updated OpenTelemetry and go-logr dependencies.
* Updated dependencies to address security vulnerabilities.
* Updated the Go version used by CI and release builds.

---
## 1.0.0 (June 14, 2021)

* Fix plugin version attribution for plugin binaries [GH-24]

## 0.0.2 (April 19, 2021)

* Include docs.zip in the release files

## 0.0.1 (April 16, 2021)

* Alicloud Plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md)
