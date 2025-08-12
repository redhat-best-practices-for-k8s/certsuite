NewCommand` – Package **version**

| Item | Detail |
|------|--------|
| **Signature** | `func NewCommand() *cobra.Command` |
| **Purpose** | Creates the *Version* sub‑command that is embedded in the top‑level `certsuite` CLI. |
| **Return value** | A fully configured `*cobra.Command` that, when invoked, prints the current build/version information of CertSuite. |

### How it works

1. **Command construction**  
   The function instantiates a new `cobra.Command` with the following fields set:
   * `Use: "version"` – the sub‑command name.
   * `Short: "Prints certsuite version"` – brief help text shown in `--help`.
   * `RunE:` a handler that writes the current build info to standard output.

2. **Build information**  
   The command pulls its data from package‑wide variables (e.g., `gitCommit`, `buildDate`) that are injected at compile time via `-ldflags`. These values are typically defined in the same file or another file in the *version* package.

3. **Global state**  
   No external global state is mutated; the only global referenced is the local variable `versionCmd` which holds the command instance for potential later use (e.g., testing). The function itself does not modify any other globals.

4. **Return value**  
   After setting up the handler, the fully configured command pointer is returned to the caller, usually the main CLI assembly routine in `cmd/certsuite/main.go`.

### Key dependencies

| Dependency | Role |
|------------|------|
| `github.com/spf13/cobra` | Provides the `Command` struct and execution framework. |
| Build‑time variables (`gitCommit`, `buildDate`) | Populate the version string displayed to the user. |

### Side effects & constraints

* **No side‑effects** beyond creating a new command instance; it does not alter global state or perform I/O.
* The function is deterministic: calling it multiple times yields distinct, independent command objects.
* It relies on compile‑time variable injection; if those variables are unset the output will show placeholders (e.g., `unknown`).

### Where it fits

The *version* package provides a dedicated CLI sub‑command.  
`NewCommand` is the single public entry point that other packages use to embed this functionality into the main command tree, keeping the version logic isolated and reusable.

---

#### Suggested Mermaid diagram (package layout)

```mermaid
graph TD;
    certsuite_main -->|addSubCmd| version_pkg[version.NewCommand()]
    subgraph version_pkg
        NewCommand() --> cmd_cobra[cobra.Command]
    end
```
This diagram shows the main CLI delegating to `NewCommand`, which returns a `cobra.Command` that is added to the command hierarchy.
