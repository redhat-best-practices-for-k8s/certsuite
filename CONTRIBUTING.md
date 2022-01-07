# Contributing

All contributions are valued and welcomed, whether they come in the form of code, documentation, ideas or discussion.
While we have not applied a formal Code of Conduct to this, and related, repositories, we require that all contributors
conduct themselves in a professional and respectful manner.

## Peer review

Although this is an open source project, an approval is required from at least two of the
[CNF Cert team members with write privileges]https://github.com/orgs/test-network-function/teams/cnfcert/members)
prior to merging a Pull Request.

*No Self Review is allowed.*  Each Pull Request will be peer reviewed prior to merge.

## Workflow

If you have a problem with the tools or want to suggest a new addition, The first thing to do is create an
[Issue](https://github.com/test-network-function/test-network-function/issues) for discussion.

When you have a change you want us to include in the main codebase, please open a
[Pull Request](https://github.com/test-network-function/test-network-function/pulls) for your changes and link it to the
associated issue(s).

### Fork and Pull

This project uses the "Fork and Pull" approach for contributions.  In short, this means that collaborators make changes
on their own fork of the repository, then create a Pull Request asking for their changes to be merged into this
repository once they meet our guidelines.

How to create and update your own fork is outside the scope of this document but there are plenty of
[more in-depth](https://gist.github.com/Chaser324/ce0505fbed06b947d962)
[instructions](https://reflectoring.io/github-fork-and-pull/) explaining how to go about this.

Once a change is implemented, tested, documented, and passing all the checks then submit a Pull Request for it to be
reviewed by the maintainers listed above.  A good Pull Request will be focussed on a single change and broken into
multiple small commits where possible.  As always, you should ensure that tests should pass prior to submitting a Pull
Request.  To run the unit tests issue the following command:

```bash
make test
```

Changes are more likely to be accepted if they are made up of small and self-contained commits, which leads on to
the next section.

### Commits

A good commit does a *single* thing, does it completely, and describes *why*.

The commit message should explain both what is being changed, and in the case of anything non-obvious why that change
was made.  Commit messages are again something that has been widely written about, so need not be discussed in detail
here.

Contributors should follow [these seven rules](https://chris.beams.io/posts/git-commit/#seven-rules) and keep individual
commits focused (`git add -p` will help with this).

## Test Implementation guidelines

Each contributed test is expected to implement the `reel.Handler` and `tnf.Tester` interfaces.  Additionally, each test
must be based on CLI commands.  No tests should utilize OpenShift client libraries.  The choice to avoid OpenShift
client is deliberate, and was decided to aid in support of all versions of OpenShift despite the API(s) changing.
Generally speaking, the CLI API changes much less quickly.

For more details on how to add a new Test either as a JSON definition, or in code, please see [DEVELOPING.md](DEVELOPING.md)

### Unit Testing Tests

Each `tnf.Tester` implementation must have unit tests.  Ideally, it should strive for 100% line coverage when possible.  For some examples of existing unit tests, consult:

* pkg/tnf/handlers/base/version_test.go
* pkg/tnf/handlers/hostname/hostname_test.go
* pkg/tnf/handlers/ipaddr/ipaddr_test.go
* pkg/tnf/handlers/ping/ping_test.go

As always, you should ensure that tests should pass prior to submitting a Pull Request.  To run the unit tests issue the
following command:

```bash
make test
```
## Configuration guidelines

Many Tests will require some form of extra configuration.  To maintain reproducibility and auditability outcomes this
configuration must be included in a claim file. For all current configuration approaches (see the `generic` test spec)
this will be done automatically provided the `config` structure for the Test implements or inherits a working `MarshalJSON` and `UnmarshalJSON`
interface so it can be included in a
[test-network-function-claim](https://github.com/test-network-function/test-network-function-claim) JSON file.

All configuration must adhere to these two requirements will automatically be included in the claim.

## Documentation guidelines

Each exported API, global variable or constant must have proper documentation which adheres to `gofmt`.

Each non-test `package` must have a package comment.  Package comments must be block comments (`/* */`), unless they are
short enough to fit on a single line when a line comment is allowed.

Changes must also include updates to affected documentation.  This means both in-code documentation and the accompanying
files such as this one.  If a change introduces a new behaviour, interface or capability then it is even more important
that the accompanying documentation and guides are updated to include that information.

## Style guidelines

Ensure `goimports` has been run against all Pull Requests prior to submission.

In addition, the `test-network-function` project committers expect all Pull Requests have no linting errors when the
configured linters are used.  Please ensure you run `make lint` and resolve any issues in your changes before submitting
your PR.  Disabled linting must be justified.

Finally, all contributions should follow the guidance of [Effective Go](https://golang.org/doc/effective_go.html)
unless there is a clear and considered reason not to.  Contributions are more likely to be accepted quickly if any
divergence from the guidelines is justified before someone has to ask about it.

## Mock guidelines

The `test-network-function` project embraces GoLang `interface`s.  As such, pull requests should contain mock
implementations whenever necessary.  If an upstream dependency does not contain an interface or mock implementation,
the code author is expected to create a minimal shim interface that represents the needed functionality.  Tests that
utilize actual implementations will not be considered for merge.  Traditionally, we suggest using `mockgen`, though
other mock generators will be considered if they provided technology above and beyond what mockgen provides.

As an example, consult `spawner.SpawnFunc`.  For unit test case contexts, Unix commands should not be executed, so a
mock implementation is provided in `mock_interactive.MockSpawnFunc`.

Mocks for `test-network-function` interfaces should be auto-generated when possible.  Please add any necessary `mockgen`
implementations to the `mocks` Makefile target.

### Do not store mock implementations in source

Mocks generate Go source files.  As such, different versions of `mockgen` compiled for various platforms are known to
output drastically different code.  `test-network-function` takes the stance that the node compiling the code should
also compile the mocks just before they are needed.  Thus, `make mocks` is part of the default `make build` target.

As such, if you decide to add a `mockgen` invocation to the `mocks` target in [Makefile](Makefile), then please ensure
you also add the `-destination` to [.gitignore](.gitignore) so it is ignored by `git`.

If you decide you want to test a given generated mock without building the code, issue the following command:

```bash
make mocks
```

### Externally generated mocks

Although `test-network-function` attempts to limit the use of third party libraries, sometimes they cannot be avoided.
For example `go-expect` is used in the implementation for the `tnf.Test` type.  The issue comes when third party
libraries fail to provide mock implementations for Go `interface`s.  In this case, you should:

1) Clone the appropriate source version.
2) Manually invoke `mockgen` on the source file containing the interface to an appropriate destination in the
`test-network-function` source tree.
3) Commit the implementation to source.
4) Add the exception to "Known manually generated mocks" table below.

An example implementation is the
[`go-expect` `expect.Expecter` interface mock](pkg/tnf/interactive/mocks/mock_expect.go).

If you need to upgrade the dependency in the future, make sure to re-apply this procedure in order to pick up any
API changes that might have occurred between versions.

#### Known manually generated mocks

Interface|Implementation
---|---
`expect.Expecter`|[mock_expect.go](pkg/tnf/interactive/mocks/mock_expect.go)
