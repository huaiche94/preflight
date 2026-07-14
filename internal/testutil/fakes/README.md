# internal/testutil/fakes/ — test doubles for the frozen app ports

> 🌐 English | [繁體中文](README.zh-TW.md)

Hand-written, configurable test doubles for the frozen cross-component
service interfaces in [../../app/ports.go](../../app/ports.go), shared so
each role does not hand-roll its own. See doc.go for the package
contract.

Every fake follows one pattern (doc.go spells it out):

- one exported struct per frozen interface, named `Fake<InterfaceName>`,
  with one optional `<Method>Func` field per method — a test configures
  exactly the methods it needs;
- a compile-time `var _ app.X = (*FakeX)(nil)` assertion, so a frozen
  contract change breaks this package at build time rather than at a
  downstream test;
- calling a method whose Func field is nil fails loud with the frozen
  `domain.Error` shape (`ErrCodeUnavailable`; unconfigured.go) instead
  of silently returning zero values;
- no built-in call recording or synchronization — a test that needs
  those builds them into its own Func closures.

Fakes provided: `FakeEvaluationService`, `FakeGracefulPauseService`,
`FakeProgressTreeService`, `FakeRepositoryCheckpointService`,
`FakeStateCheckpointService`, `FakeTurnInterrupter`,
`FakeSessionResumer`.

providercontract.go additionally exports a reusable contract-test suite
(`ProviderInterrupterContract`, `ProviderSessionResumerContract`), run
here against the two provider fakes (providercontract_test.go) and
intended for any future real interrupter/resumer adapter.
