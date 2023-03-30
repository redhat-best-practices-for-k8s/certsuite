<!-- markdownlint-disable line-length -->
# Contributing

All contributions are valued and welcomed, whether they come in the form of code, documentation, ideas or discussion.
While we have not applied a formal Code of Conduct to this, and related, repositories, we require that all contributors
conduct themselves in a professional and respectful manner.

## Peer review

Although this is an open source project, an approval is required from at least two of the
[CNF Cert team members with write privileges](https://github.com/orgs/test-network-function/teams/cnfcert/members)
prior to merging a Pull Request.

*No Self Review is allowed.* Each Pull Request will be peer reviewed prior to merge.

## Workflow

If you have a problem with the tools or want to suggest a new addition, the first thing to do is create an
[Issue](https://github.com/test-network-function/cnf-certification-test/issues) for discussion.

When you have a change you want us to include in the main codebase, please open a
[Pull Request](https://github.com/test-network-function/cnf-certification-test/pulls) for your changes and link it to the
associated issue(s).

### Fork and Pull

This project uses the "Fork and Pull" approach for contributions. In short, this means that collaborators make changes
on their own fork of the repository, then create a Pull Request asking for their changes to be merged into this
repository once they meet our guidelines.

How to create and update your own fork is outside the scope of this document but there are plenty of
[more in-depth](https://gist.github.com/Chaser324/ce0505fbed06b947d962)
[instructions](https://reflectoring.io/github-fork-and-pull/) explaining how to go about this.

Once a change is implemented, tested, documented, and passing all the checks then submit a Pull Request for it to be
reviewed by the maintainers listed above. A good Pull Request will be focused on a single change and broken into
multiple small commits where possible. As always, you should ensure that tests should pass prior to submitting a Pull
Request. To run the unit tests issue the following command:

```bash
make test
```

Changes are more likely to be accepted if they are made up of small and self-contained commits, which leads on to
the next section.

### Commits

A good commit does a *single* thing, does it completely, and describes *why*.

The commit message should explain both what is being changed, and in the case of anything non-obvious why that change
was made. Commit messages are again something that has been widely written about, so need not be discussed in detail
here.

Contributors should follow [these seven rules](https://chris.beams.io/posts/git-commit/#seven-rules) and keep individual
commits focused (`git add -p` will help with this).

### Unit Testing Tests

Each `tnf.Tester` implementation must have unit tests. Ideally, it should strive for 100% line coverage when possible. For some examples of existing unit tests, consult:

* pkg/tnf/handlers/base/version_test.go
* pkg/tnf/handlers/hostname/hostname_test.go
* pkg/tnf/handlers/ipaddr/ipaddr_test.go
* pkg/tnf/handlers/ping/ping_test.go

As always, you should ensure that tests should pass prior to submitting a Pull Request. To run the unit tests issue the
following command:

```bash
make test
```

## Configuration guidelines

Many Tests will require some form of extra configuration. To maintain reproducibility and auditability outcomes this
configuration must be included in a claim file. For all current configuration approaches (see the `generic` test spec)
this will be done automatically provided the `config` structure for the Test implements or inherits a working `MarshalJSON` and `UnmarshalJSON`
interface so it can be included in a
[test-network-function-claim](https://github.com/test-network-function/test-network-function-claim) JSON file.

All configuration must adhere to these two requirements will automatically be included in the claim.

## Documentation guidelines

Each exported API, global variable or constant must have proper documentation which adheres to `gofmt`.

Each non-test `package` must have a package comment. Package comments must be block comments (`/* */`), unless they are
short enough to fit on a single line when a line comment is allowed.

Changes must also include updates to affected documentation. This means both in-code documentation and the accompanying
files such as this one. If a change introduces a new behaviour, interface or capability then it is even more important
that the accompanying documentation and guides are updated to include that information.

## Style guidelines

Ensure `goimports` has been run against all Pull Requests prior to submission.

In addition, the `test-network-function` project committers expect all Pull Requests have no linting errors when the
configured linters are used. Please ensure you run `make lint` and resolve any issues in your changes before submitting
your PR. Disabled linting must be justified.

Finally, all contributions should follow the guidance of [Effective Go](https://golang.org/doc/effective_go.html)
unless there is a clear and considered reason not to. Contributions are more likely to be accepted quickly if any
divergence from the guidelines is justified before someone has to ask about it.
