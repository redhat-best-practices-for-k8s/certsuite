IsRHEL` – Detect Red‑Hat Enterprise Linux

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/isredhat` |
| **Exported?** | Yes (`func IsRHEL(string) bool`) |
| **Signature** | `func IsRHEL(osRelease string) bool` |
| **Purpose** | Inspect the contents of an operating‑system release file (normally `/etc/os-release`) and decide whether it represents a Red‑Hat Enterprise Linux (RHEL) distribution. |

---

### How It Works

```go
func IsRHEL(s string) bool {
    // 1️⃣ Match non‑Red‑Hat identifiers
    notRH := MustCompile(NotRedHatBasedRegex)
    if len(notRH.FindAllString(s, -1)) > 0 {
        return false
    }

    // 2️⃣ Look for a RHEL‑specific version string
    rhelVer := MustCompile(VersionRegex)
    if len(rhelVer.FindAllString(s, -1)) == 0 {
        Info("RHEL not found")
        return false
    }
    return true
}
```

| Step | Action | Dependency |
|------|--------|------------|
| **1️⃣** | Compile `NotRedHatBasedRegex` and search the input. If any match is found, the OS cannot be RHEL. | `regexp.MustCompile`, `FindAllString` |
| **2️⃣** | Compile `VersionRegex` (the pattern that matches a RHEL version string) and confirm at least one occurrence. If none, log via `Info`. | `regexp.MustCompile`, `FindAllString`, `len`, `Info` |

The function returns **`true` only when the input contains no Red‑Hat‑excluded markers *and* does contain a RHEL‑style version string**.

---

### Inputs / Outputs

| Parameter | Type | Meaning |
|-----------|------|---------|
| `s` | `string` | Raw text from `/etc/os-release` (or any similar file). |

| Return | Type | Meaning |
|--------|------|---------|
| `bool` | `true` if the OS is RHEL; `false` otherwise. |

---

### Dependencies

- **Regular expressions**  
  - `NotRedHatBasedRegex`: pattern that matches identifiers for non‑RHEL distributions (e.g., "Alpine", "Ubuntu").  
  - `VersionRegex`: pattern that captures a valid RHEL version string like `"VERSION_ID=\"8.4\""`.  

- **Logging** – Calls `Info` to emit a message when no RHEL signature is found.

---

### Side‑Effects

- No state mutation; the function is pure aside from logging.
- Logging may produce console output depending on the `Info` implementation in the test harness.

---

### Package Context

The `isredhat` package contains utilities used by certsuite tests to determine platform type.  
`IsRHEL` is a small helper that enables other tests (e.g., for RHEL‑specific configurations) to gate their logic based on whether the node under test runs RHEL.

```mermaid
flowchart TD
    A[Input string] --> B{Match NotRedHatBasedRegex?}
    B -- Yes --> C[Return false]
    B -- No --> D{Match VersionRegex?}
    D -- No --> E[Info("RHEL not found")]
    E --> F[Return false]
    D -- Yes --> G[Return true]
```

---

### Summary

`IsRHEL` is a concise, read‑only routine that parses an OS release file for Red‑Hat specific markers and returns a boolean flag. It serves as the foundational check for other certsuite tests that need to operate only on RHEL systems.
