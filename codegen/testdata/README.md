# Compatibility fixtures

`counter.json`, `example-counter.json`, `stream.json`, and `bsim-auth.json`
freeze validated pluginapi models, ordered package files and renderer errors.
Supported file bytes were checked against an independently built signed core
`v0.4.5` CLI (commit `cfed4182356a4f4609c88f58924aac79c05ae5b6`). File order
and backend refusal text come from the landed donor adapter at
`ef6e04bfae8dab0f8e1ac40c9bb10713dbf09250`.

Each JSON records the input SHA256. The immutable bsim-auth input is B6 commit
`2d23abdafa1e0f68c6003ab56274b2ac38378ef9`, SHA256
`1be1ed52ac9a85a62d5c5e371a9e22d38f834282681528476ca071e7bfc2cb66`.
`CounterClient.swift.golden` preserves the donor's original renderer regression.

`TestPluginReleasedBaselineParity` drives `Plugin.Generate` and compares all
ordered filenames, bytes and error text. The independent facade CLI rejects
stream schemas before invoking Swift with exit 2 and its existing diagnostic;
the backend rejects them with its existing unsupported-field diagnostic and
returns no files. These are distinct preserved entry-point contracts.

The fixture set bounds compatibility evidence; it does not enumerate every
possible valid schema. `TestPluginRejectsStreamsWithoutPartialFiles` extends
negative coverage to request/response and mixed payloads. Generated native
execution in `TestGeneratedSwiftFixedBytesRejectBeforeTransmit` checks real
Swift fixed-byte guards, supplementing preserved source-snippet tests.
