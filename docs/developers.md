<!-- markdownlint-disable line-length no-bare-urls -->
# Steps

To test the newly added test / existing tests locally, follow the steps

- Clone the repo
- Set runtime environment variables, as per the requirement.

    For example, to deploy partner deployments in a custom namespace in the test config.

    ```yaml
    targetNameSpaces:
      - name: mynamespace
    ```

- Also, skip intrusive tests

```shell
export CERTSUITE_NON_INTRUSIVE_ONLY=true
```

- Set K8s config of the cluster where test pods are running

    ```shell
    export KUBECONFIG=<<mypath/.kube/config>>
    ```

- Execute test suite, which would build and run the suite

    For example, to run `networking` tests

    ```shell
    ./certsuite run -l networking
    ```

# Dependencies on other PR

If you have dependencies on other Pull Requests, you can add a comment like that:

```text
Depends-On: <url of the PR>
```

and the dependent PR will automatically be extracted and injected in your change during the GitHub Action CI jobs and the DCI jobs.

# Linters for the Codebase

- [`checkmake`](https://github.com/mrtazz/checkmake)
- [`golangci-lint`](https://github.com/golangci/golangci-lint)
- [`hadolint`](https://github.com/hadolint/hadolint)
- [`markdownlint`](https://github.com/igorshubovych/markdownlint-cli)
- [`shellcheck`](https://github.com/koalaman/shellcheck)
- [`shfmt`](https://github.com/mvdan/sh)
- [`typos`](https://github.com/crate-ci/typos)
- [`yamllint`](https://github.com/adrienverge/yamllint)
