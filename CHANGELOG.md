# Changelog

## [1.1.0](https://github.com/meigma/blob-cli/compare/v1.0.0...v1.1.0) (2026-01-25)


### Features

* **cache:** implement comprehensive caching support ([#27](https://github.com/meigma/blob-cli/issues/27)) ([ed360f2](https://github.com/meigma/blob-cli/commit/ed360f2cf395158d0cf18a81eb1a2d79cebcca2d))
* **ci:** add publish-src workflow for test image ([#23](https://github.com/meigma/blob-cli/issues/23)) ([b96d513](https://github.com/meigma/blob-cli/commit/b96d5130506dc9e148ed410a20649ded397d75a8))


### Bug Fixes

* **deps:** update blob to v1.1.1 ([#28](https://github.com/meigma/blob-cli/issues/28)) ([ccec3ee](https://github.com/meigma/blob-cli/commit/ccec3ee631c979d9993e6f42a3d44aaa89620e11))
* **policy:** use branch name without refs/heads prefix ([#25](https://github.com/meigma/blob-cli/issues/25)) ([78d6aac](https://github.com/meigma/blob-cli/commit/78d6aac709f2a056f607cc00c503c44f993bd220))

## 1.0.0 (2026-01-24)


### Features

* add CLI boilerplate with Cobra/Viper ([#1](https://github.com/meigma/blob-cli/issues/1)) ([1e2b11f](https://github.com/meigma/blob-cli/commit/1e2b11f0b9eb8274edcb5033a37d70ddc5ab0765))
* add configuration foundation with typed config and context propagation ([#5](https://github.com/meigma/blob-cli/issues/5)) ([f1b0167](https://github.com/meigma/blob-cli/commit/f1b0167470daef656541e0bfc0998d943d7c1f70))
* **alias:** implement alias list, set, and remove commands ([#7](https://github.com/meigma/blob-cli/issues/7)) ([efa1865](https://github.com/meigma/blob-cli/commit/efa186562db2fb13246de7d618750006f298b21a))
* **cat,cp:** implement cat and cp commands ([#13](https://github.com/meigma/blob-cli/issues/13)) ([3e60c94](https://github.com/meigma/blob-cli/commit/3e60c94ab889dbe3989fd95cc6c4dcb836943f3a))
* **deps:** upgrade blob library to v1.1.0 ([#14](https://github.com/meigma/blob-cli/issues/14)) ([e5c5e8b](https://github.com/meigma/blob-cli/commit/e5c5e8b77786e3bbcbfc8a01f11ba783c4518b98))
* **inspect:** implement inspect command ([#15](https://github.com/meigma/blob-cli/issues/15)) ([2b51a30](https://github.com/meigma/blob-cli/commit/2b51a30f70721d96679eb694391d54aa175313c7))
* **ls,tree:** implement ls and tree commands ([#9](https://github.com/meigma/blob-cli/issues/9)) ([4c40d6e](https://github.com/meigma/blob-cli/commit/4c40d6e9cdeb86434e37c3db811bc5b1414e9527))
* **open:** implement interactive TUI file browser ([#19](https://github.com/meigma/blob-cli/issues/19)) ([734e835](https://github.com/meigma/blob-cli/commit/734e83528cc81db0cf03491d7c1c309446fae584))
* **pull:** implement pull command with policy verification ([#11](https://github.com/meigma/blob-cli/issues/11)) ([0d98c25](https://github.com/meigma/blob-cli/commit/0d98c2511db06233abc8d4b803ac46579003f2ed))
* **push:** implement push command with signing support ([#8](https://github.com/meigma/blob-cli/issues/8)) ([8d269cf](https://github.com/meigma/blob-cli/commit/8d269cf3f43b5325429e286161d4bba3e811e276))
* **release:** add release automation with goreleaser and release-please ([#20](https://github.com/meigma/blob-cli/issues/20)) ([c370975](https://github.com/meigma/blob-cli/commit/c370975087bb3d9f56a08b3d657943513f094abb))
* **sign,verify:** implement sign and verify commands ([#16](https://github.com/meigma/blob-cli/issues/16)) ([61a4519](https://github.com/meigma/blob-cli/commit/61a45198922d39cd157cf5255b2014bd688a5d76))
* **tag:** implement tag command ([#17](https://github.com/meigma/blob-cli/issues/17)) ([79a5800](https://github.com/meigma/blob-cli/commit/79a5800654446586fa16cd618c2297dd25b1d122))
* **test:** add comprehensive integration tests ([#18](https://github.com/meigma/blob-cli/issues/18)) ([856b936](https://github.com/meigma/blob-cli/commit/856b93645ca4b188dcb45283f9b0b2c37d33e354))


### Code Refactoring

* **tests:** use testify assert/require for config tests ([#6](https://github.com/meigma/blob-cli/issues/6)) ([9222a5e](https://github.com/meigma/blob-cli/commit/9222a5e48e03e6ce8d03bf4bd402cea4d2ccb698))

## Changelog
