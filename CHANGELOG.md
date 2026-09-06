# Changelog

All notable changes to this project will be documented in this file.

## [0.57.6](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.57.5...v0.57.6) (2026-09-06)

### Chores

* require go 1.27 ([8cd463b](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/8cd463bc664792f0a367d0694a199e11ec27f61f))
* update dependencies ([020c9e9](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/020c9e94672a0dea34f7930d270c07e771800cfe))

## [0.57.5](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.57.4...v0.57.5) (2026-09-06)

### Bug Fixes

* update gosec and govulncheck to latest tooling versions ([4f804b8](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/4f804b827f6053e1e97bf9f335d5d56eaee4c7d6))

## [0.57.4](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.57.3...v0.57.4) (2026-09-06)

### Bug Fixes

* correct teams datasource schema mapping and handle not found errors ([899add5](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/899add5439c9ea75691805fb4f47cfdb893dc0c2))
* correct validation logic, descriptions and regexes across validators ([3302fe5](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/3302fe540b4ac958f48ce5cb60f3131ffba50393))
* enforce non-empty ports in range validator and align DH error format ([e991f64](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/e991f64a3f680ad98063c7dbfa2250008d2b8d5a))
* handle 404 not found in resource read and delete operations ([66c2a0b](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/66c2a0bed6e896560ac92d937eb497399224129c))
* improve match detection and nil safety in subnet datasource ([0aaa7e9](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/0aaa7e9d2a1bde14454cab6b37bc99b7139dca76))
* propagate errors on optional parent resource lookup failures ([4f8a515](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/4f8a515097b75e6bdff7418ae8f7a9890573e8dc))
* synchronize state from API response during resource update ([70fd9ec](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/70fd9ec031b78124dd1d4db78e75e5a58eafa6bd))
* use fast syntax validation for email validator ([c04c20c](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/c04c20caea6a50ec1939ca626d48b10b56f01d55))

## [0.57.3](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.57.2...v0.57.3) (2026-09-06)

### Chores

* update go-git and circl dependencies ([6b963c4](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/6b963c4ec4373d79996f136e93ce31787f1f6f5d))

## [0.57.2](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.57.1...v0.57.2) (2026-09-06)

### Chores

* upgrade terraform-plugin-docs and transitive dependencies ([404811e](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/404811e2cf974edd0986b4edd46088a039f1c644))

## [0.57.1](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.57.0...v0.57.1) (2026-09-06)

### Bug Fixes

* correct ipsec connection resource assertion, defaults and schema descriptions ([9ad9d44](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/9ad9d4407182314a74d3b6d49e39c558f2728a49))
* correct typos in resource schema descriptions and docs ([1bfdec4](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/1bfdec4bc3c3e388141573db65b92e5d43cee603))
* guard against nil pointers and empty identifiers in resource lookup helpers ([1415db5](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/1415db549fcd9f7c435812cafb237f86476db771))
* populate computed subnet state attributes on resource creation ([8ce55c4](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/8ce55c4312c724daaeb5a452eda321d1999d6fb3))
* populate state and guard against nil in dns record resource ([caf89f9](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/caf89f944d22edc0145a95821f3b6a17ef97a62d))
* prevent nil dereferences in project quotas and private subnets mapping ([eed1526](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/eed1526bd83165a572dbeacb2ae80159ca62a0c3))

### Chores

* modernize makefile tooling and lint targets ([23deebb](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/23deebb7fb081835bddbd7f2ef313e9d7ef67249))
* update dependencies to resolve security vulnerabilities ([3c6d83d](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/3c6d83d703319ec4dc77b94dc8078b1e494a5a02))

### Performance Improvements

* avoid unnecessary API lookups for optional pool, template and nfs parameters ([a1fcdfc](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/a1fcdfcde2ffb376194b72dbfca2df9ee61d4fd7))

## [0.57.0](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/compare/v0.56.1...v0.57.0) (2026-09-06)

### Chores

* update kowabunga-go SDK to v0.54.0 ([882aa02](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/882aa02c3ff0f481c15add16b6298f55e4869f58))
* update release workflow to latest actions and conventional commits ([cc7c9e1](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/cc7c9e1a248a3d99ffa81136d544ec1a583b20d1))

### Features

* add uefi support to instance and kompute resources ([0331b32](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/commit/0331b3299c200d48ce8eca0ac20555c3be74cf13))

# Changelog

## 0.56.1 (2025-10-26)

The Kowabunga project team is happy to announce the immediate availability of Kowabunga Terraform provider v0.56.1.

* **NEW**: implement API v0.53.2 changes (domain attribute for region).
* **BUG**: fix **kowabunga_dns_record** read back.

## 0.56.0 (2025-10-18)

The Kowabunga project team is happy to announce the immediate availability of Kowabunga Terraform provider v0.56.0.

* **NEW**: updated Go modules dependencies.
* **NEW**: implement API v0.53.1 changes (DNS record for region).

## 0.55.1 (2025-09-11)

The Kowabunga project team is happy to announce the immediate availability of Kowabunga Terraform provider v0.55.1.

* **NEW**: updated Go modules dependencies.
* **NEW**: updated Go compiler to 1.25.
* **BUG**: Fix argument passing in **kawaii** resource creation.

## 0.55.0

The Kowabunga project team is happy to announce the immediate availability of Kowabunga Terraform provider v0.55.0.
