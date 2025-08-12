configOption`

| Item | Detail |
|------|--------|
| **File** | `cmd/certsuite/generate/config/config.go` (line 18) |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config` |
| **Exported?** | No – the type is unexported (`configOption`). It is used only within this package. |

### Purpose

`configOption` represents a single configuration flag that can be supplied to the CertSuite generator CLI.  
The struct holds two pieces of data:

* `Help`: A short description shown when the user asks for help or documentation.  
* `Option`: The actual command‑line switch (e.g., `"--skip-tests"`).

This design keeps a clean separation between what the user sees and how the program internally parses options.

### Fields

| Field | Type   | Role |
|-------|--------|------|
| `Help`  | `string` | Human‑readable description of the option. Used by the help formatter to inform users. |
| `Option`| `string` | The raw flag string that will be passed to the CLI parser (e.g., `"--config"`). |

### Typical Usage

```go
// Inside generate/config/config.go
var options = []configOption{
    {"Generate a fresh cert bundle", "--generate"},
    {"Skip existing files", "--skip-existing"},
}
```

When building the command‑line interface, these structs are iterated over to:

1. Register each `Option` with a flag parser (e.g., `flag.BoolVar`).  
2. Build the help text from `Help`.

### Dependencies & Side Effects

* **Dependencies** – None beyond the Go standard library (`fmt`, `flag`, etc.).  
* **Side Effects** – Instantiating a `configOption` has no side effects; it is plain data.

### Integration into the Package

The package `generate/config` centralizes all command‑line configuration for CertSuite’s generate tool.  
`configOption` serves as the building block:

1. A slice of these structs defines every flag that the CLI supports.
2. The generator logic later reads the parsed flag values (stored elsewhere) to decide behavior.

Because the type is unexported, other packages cannot directly create or manipulate `configOption`s; they must go through the package’s exported APIs (`GenerateConfig`, etc.). This encapsulation ensures consistency and prevents accidental misuse of raw option strings.
