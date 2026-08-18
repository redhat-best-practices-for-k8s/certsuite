# Share Test Suite Results with Red Hat

Your test suite results can be collected and stored in our
[Collector](https://github.com/redhat-best-practices-for-k8s/collector) database.

## What information is shared

* Only partner name and pass/fail/skip status per test case is stored
* Protected by password, partners can access only the data they have submitted.
Red Hat team can access data from all partners.

## Why should I store my data in the Collector?

* Keep track of your test suite results over time.
* Contribute to our statistics and analysis,
to improve Red Hat best practices test suite for Kubernetes.

## How to enable collection

Pass `--enable-data-collection` to `certsuite run` and set these fields in
`certsuite_config.yml` (see [configuration](configuration.md)):

```yaml
executedBy: ""
partnerName: ""
collectorAppPassword: ""
collectorAppEndpoint: "http://claims-collector.cnf-certifications.sysdeseng.com"
```

To upload artifacts to the Red Hat Connect portal instead, set `connectAPIConfig`
(or the `--connect-api-*` flags) with a project ID and API key.
