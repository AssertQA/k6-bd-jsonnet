# k6 Jsonnet Extension

Generate JSON payloads from Jsonnet templates inside k6 scripts, with optional
fake data generation via `gofakeit`. This extension exposes a small JavaScript
API that lets you load and evaluate Jsonnet files at runtime and toggle test
data generation with a single flag.

## What this provides

- Jsonnet evaluation using `go-jsonnet`
- Optional fake data generation via a `fake()` Jsonnet native function
- A simple k6 JS API for processing templates or reading raw template text
- Support for a `generateTestData` external variable inside Jsonnet

## Install

Build a custom k6 binary with xk6:

```bash
xk6 build --with github.com/AssertQA/k6-bd-jsonnet
```

This produces a `k6` binary in the current directory.

## JS API

Import the module in your k6 script:

```javascript
import jsonnet from "k6/x/jsonnet";
```

### `processTemplate(templatePath, generateTestData)`

Evaluate a Jsonnet file and return the rendered JSON string.

- `templatePath`: string path to the Jsonnet file
- `generateTestData`: boolean; when `true`, a `fake()` function is available in
  Jsonnet and the `generateTestData` external variable is set to `true`

Returns a JSON string.

### `generateTestData(templatePath)`

Convenience wrapper for `processTemplate(templatePath, true)`.

### `loadTemplate(templatePath)`

Loads a template file as plain text without evaluating it.

Returns the file contents as a string.

## Jsonnet features and conventions

### External variable: `generateTestData`

The extension injects `generateTestData` as a Jsonnet external variable. In
Jsonnet, access it using:

```jsonnet
std.extVar("generateTestData")
```

Use this to switch between realistic fake data and fixed/static values.

### Native function: `fake(pattern)`

When `generateTestData` is `true`, a Jsonnet native function `fake()` is
registered. It is backed by `gofakeit.Generate`, which accepts templated
patterns such as `"{firstname}"`, `"{email}"`, or `"{number:1,100}"`.

When `generateTestData` is `false`, `fake()` is not registered; attempting to
call it will fail during Jsonnet evaluation.

## Example: Jsonnet template

`templates/user.jsonnet`:

```jsonnet
{
  id: std.extVar("generateTestData")
    ? fake("{number:1000,9999}")
    : "1001",
  name: std.extVar("generateTestData")
    ? fake("{firstname} {lastname}")
    : "Jane Doe",
  email: std.extVar("generateTestData")
    ? fake("{email}")
    : "jane@example.com",
  createdAt: std.extVar("generateTestData")
    ? fake("{date}")
    : "2024-01-01",
}
```

## Example: k6 script

```javascript
import http from "k6/http";
import jsonnet from "k6/x/jsonnet";

export default function () {
  const json = jsonnet.processTemplate(
    "./templates/user.jsonnet",
    true
  );

  // Convert to object for request body
  const payload = JSON.parse(json);

  http.post("https://httpbin.test.k6.io/post", JSON.stringify(payload), {
    headers: { "Content-Type": "application/json" },
  });
}
```

## Example: read raw template text

```javascript
import jsonnet from "k6/x/jsonnet";

export default function () {
  const raw = jsonnet.loadTemplate("./templates/user.jsonnet");
  console.log(raw);
}
```

## Notes and limitations

- Templates are evaluated by file path; ensure Jsonnet imports are relative to
  the template directory or use absolute paths.
- `fake()` is only available when `generateTestData` is `true`.
- Errors during evaluation are logged and return an empty string.

## Development

To run locally, build with xk6 as shown above and run your script with the
custom `k6` binary.