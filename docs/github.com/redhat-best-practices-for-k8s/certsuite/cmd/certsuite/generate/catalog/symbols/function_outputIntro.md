outputIntro` – Package‑level helper for the *generate catalog* command

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/catalog` |
| **Signature** | `func() string` |
| **Exported?** | No – it is an internal helper. |

### Purpose

`outputIntro` builds the introductory text that appears at the top of the Markdown file produced by the `generate catalog` sub‑command.  
It simply returns a fixed string, but keeping this logic in its own function makes the command implementation cleaner and isolates future changes (e.g., adding version numbers or build metadata).

### Inputs / Outputs

- **Inputs**: none – the function takes no parameters.
- **Output**: a `string` containing Markdown markup.

The returned string is concatenated with the rest of the catalog content before being written to disk by the command handler.

### Dependencies & Side‑Effects

| Dependency | Description |
|------------|-------------|
| None (no imports or global reads) | The function contains only a literal return value. |
| Global variables (unused) | `generateCmd`, `markdownGenerateClassification`, and `markdownGenerateCmd` are defined in the same file but *not* referenced here, so they have no effect on this function. |

Because it performs no I/O or state mutation, calling `outputIntro` is completely side‑effect free.

### Role in the Package

The `generate catalog` command constructs a Markdown document that documents the certificate catalog.  
`outputIntro` supplies the leading section (typically a title and brief description). The rest of the command (`generateCmd`, `markdownGenerateClassification`, etc.) appends the classification tables, detailed entries, and footer.

**Diagram – call flow**

```mermaid
graph TD
  A[User runs `certsuite generate catalog`] --> B[Command handler]
  B --> C{build intro}
  C --> D[outputIntro()]
  D --> E[intro string]
  E --> F[Append to Markdown buffer]
  F --> G[Write file]
```

### Summary

`outputIntro` is a tiny, pure function that returns the introductory Markdown for the catalog generation process.  
It has no external dependencies or side‑effects and fits neatly into the command’s pipeline by supplying the first part of the output document.
