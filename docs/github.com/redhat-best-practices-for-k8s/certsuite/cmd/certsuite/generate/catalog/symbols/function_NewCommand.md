NewCommand` – Main entry point for the *certsuite* catalog generator

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/catalog` |
| **Signature** | `func NewCommand() *cobra.Command` |
| **Exported?** | Yes – it is the function that callers use to obtain the root command for the “catalog” sub‑command of the CertSuite CLI. |

### Purpose
`NewCommand` builds and returns a fully‑configured `*cobra.Command` that represents the top‑level *catalog* command in the CertSuite tool.  
The command tree it creates consists of:

1. The **root** command (`generateCmd`) – this is defined elsewhere in the same file (line 43). It usually contains flags and help text common to all generate sub‑commands.
2. Two child commands added by successive calls to `AddCommand`:
   * One that produces a Markdown‑formatted classification of certificates (created via `markdownGenerateClassification`).  
   * Another that generates a full Markdown catalog (`markdownGenerateCmd`).

These children are wired into the root command so that executing:

```bash
certsuite generate catalog markdown-classification ...
```

or

```bash
certsuite generate catalog markdown ...
```

dispatches to the appropriate sub‑command logic.

### Inputs / Outputs
| Parameter | Type | Description |
|-----------|------|-------------|
| None | – | The function does not take any arguments. |

| Return value | Type | Description |
|--------------|------|-------------|
| `*cobra.Command` | Pointer to a Cobra command | A fully constructed command tree ready for use by the CLI entry point. |

### Key Dependencies
- **Cobra** (`github.com/spf13/cobra`) – the library used for building CLI commands.
- Global variables defined in the same file:
  - `generateCmd`: the root *catalog* command object.
  - `markdownGenerateClassification`: a Cobra command that outputs classification data in Markdown.
  - `markdownGenerateCmd`: a Cobra command that outputs the full catalog in Markdown.

These globals are initialized earlier in the file (lines 43, 48, and 55). `NewCommand` simply attaches them to the root command via `AddCommand`.

### Side Effects
- **No external state changes** beyond attaching sub‑commands to the root.  
- The function does not read or modify files; it only configures command objects.

### How It Fits Into the Package
The *catalog* package is responsible for generating certificate catalogs in various formats. `NewCommand` is the public entry point that other parts of the CertSuite CLI (specifically, the main `certsuite` binary) call to integrate the catalog generation features into the overall command hierarchy.

```mermaid
graph TD;
    subgraph CertSuite CLI
        A[certsuite] --> B[generate]
        B --> C[catalog] --> D[NewCommand()]
        D --> E[markdownGenerateClassification]
        D --> F[markdownGenerateCmd]
    end
```

- `certsuite` is the root executable.
- `generate` groups all generation commands.
- `catalog` is the sub‑package providing catalog‑specific commands.
- `NewCommand()` supplies the Cobra command tree that the CLI registers.

Thus, `NewCommand` is the glue that connects the *catalog* functionality to the rest of the CertSuite tool.
