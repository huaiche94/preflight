# internal/testutil/ — shared test-support code

> 🌐 English | [繁體中文](README.zh-TW.md)

Umbrella directory for test-support packages shared across roles. It
holds no Go files of its own and no doc.go; the only package today is
[fakes/](fakes/) — hand-written, configurable test doubles for the
frozen cross-component service interfaces in
[../app/ports.go](../app/ports.go).

No production code imports anything under this directory; consumers are
package tests across internal/ (e.g. internal/app/wiring, internal/pause,
internal/managed) and the integration suites in
[../integrationtest/](../integrationtest/).

See [fakes/README.md](fakes/README.md) — and fakes/doc.go for the package
contract — for the fake pattern and its guarantees.
