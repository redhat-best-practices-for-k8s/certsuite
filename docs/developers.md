<!-- markdownlint-disable line-length no-bare-urls -->
# Steps

To test a newly added test or existing tests locally, follow the steps

- Clone the repo
- Build the certsuite binary

    ```shell
    make build
    ```

- Create or edit `config/certsuite_config.yml`. For example, to deploy partner
  workloads in a custom namespace:

    ```yaml
    targetNameSpaces:
      - name: mynamespace
    ```

- Skip intrusive tests if you do not want scaling or pod-recreation checks

    ```shell
    ./certsuite run --intrusive=false
    ```

- Set K8s config of the cluster where test pods are running

    ```shell
    export KUBECONFIG=<<mypath/.kube/config>>
    ```

- Execute the test suite. For example, to run `networking` tests:

    ```shell
    ./certsuite run -l networking --config-file=config/certsuite_config.yml
    ```

- List matching test cases without running them:

    ```shell
    ./certsuite info -t networking --list
    ```

- Run unit tests and linters before opening a pull request:

    ```shell
    make test
    make lint
    ```

See [Runtime environment variables](runtime-env.md) and
[Test Configuration](configuration.md) for more options.

## Dependencies on other PR

If you have dependencies on other Pull Requests, you can add a comment like that:

```text
Depends-On: <url of the PR>
```

and the dependent PR will automatically be extracted and injected in your change during the GitHub Action CI jobs and the DCI jobs.

## Linters for the Codebase

- [`checkmake`](https://github.com/mrtazz/checkmake)
- [`golangci-lint`](https://github.com/golangci/golangci-lint)
- [`hadolint`](https://github.com/hadolint/hadolint)
- [`markdownlint`](https://github.com/igorshubovych/markdownlint-cli)
- [`shellcheck`](https://github.com/koalaman/shellcheck)
- [`shfmt`](https://github.com/mvdan/sh)
- [`typos`](https://github.com/crate-ci/typos)
- [`yamllint`](https://github.com/adrienverge/yamllint)
