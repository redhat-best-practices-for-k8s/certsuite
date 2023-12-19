<!-- markdownlint-disable line-length no-bare-urls no-emphasis-as-heading -->
# Test Output

## Claim File

The test suite generates an output file, named **claim** file. This file is considered as the proof of CNFs test run, evaluated by Red Hat when **certified** status is considered.

This file describes the following

* The system(s) under test
* The tests that are executed
* The outcome of the executed / skipped tests

**Files that need to be submitted for certification**

When submitting results back to Red Hat for certification, please include the above mentioned claim file, the JUnit file, and any available console logs.

**How to add a CNF platform test result to the existing claim file?**

```go
go run cmd/tools/cmd/main.go claim-add --claimfile=claim.json
--reportdir=/home/$USER/reports
```

 **Args**:
`--claimfile is an existing claim.json file`
`--repordir :path to test results that you want to include.`

 The tests result files from the given report dir will be appended under the result section of the claim file using file name as the key/value pair.
 The tool will ignore the test result, if the key name is already present under result section of the claim file.

```json
 "results": {
 "cnf-certification-tests_junit": {
 "testsuite": {
 "-errors": "0",
 "-failures": "2",
 "-name": "CNF Certification Test Suite",
 "-tests": "14",
 ...
```

**Reference**

For more details on the contents of the claim file

* [Guide](https://redhat-connect.gitbook.io/openshift-badges/badges/cloud-native-network-functions-cnf).

## Execution logs

The test suite also saves a copy of the execution logs at [test output directory]/cnf-certsuite.log

## Results artifacts zip file

After running all the test cases, a compressed file will be created with all the results files and web artifacts to review them. The file has a UTC date-time prefix and looks like this:

20230620-110654-cnf-test-results.tar.gz

The "20230620-110654" sample prefix means "June-20th 2023, 11:06:54"

This is the content of the tar.gz file:

* claim.json
* cnf-certification-tests_junit.xml (Only if enabled via `TNF_ENABLE_XML_CREATION` environment variable)
* claimjson.js
* classification.js
* results.html

This file serves two different purposes:

1. Make it easier to store and send the test results for review.
2. View the results in the html web page. In addition, the web page (either results-embed.html or results.html) has a selector for workload type and allows the partner to introduce feedback for each of the failing test cases for later review from Red Hat. It's important to note that this web page needs the `claimjson.js` and `classification.js` files to be in the same folder as the html files to work properly.

## Show Results after running the test code

A standalone HTML page is available to decode the results.
For more details, see:
https://github.com/test-network-function/parser

## Compare claim files from two different CNF Certification Suite runs

Partners can use the `tnf claim compare` tool in order to compare two claim files. The differences are shown in a table per section.
This tool can be helpful when the result of some test cases is different between two (consecutive) runs, as it shows
configuration differences in both the CNF Cert Suite config and the cluster nodes that could be the root cause for
some of the test cases results discrepancy.

All the compared sections, except the test cases results are compared blindly, traversing the whole json tree and
sub-trees to get a list of all the fields and their values. Three tables are shown:

* Differences: same fields with different values.
* Fields in claim 1 only: json fields in claim file 1 that don't exist in claim 2.
* Fields in claim 2 only: json fields in claim file 2 that don't exist in claim 1.

Let's say one of the nodes of the claim.json file contains this struct:

```json
{
  "field1": "value1",
  "field2": {
    "field3": "value2",
    "field4": {
      "field5": "value3",
      "field6": "value4"
    }
  }
}
```

When parsing that json struct fields, it will produce a list of fields like this:

```console
/field1=value1
/field2/field3=value2
/field2/field4/field5=value3
/field2/field4/field6=finalvalue2
```

Once this list of field's path+value strings has been obtained from both claim files,
it is compared in order to find the differences or the fields that only exist on each file.

This is a fake example of a node "clus0-0" whose first CNI (index 0) has a different cniVersion
and the ipMask flag of its first plugin (also index 0) has changed to false in the second run.
Also, the plugin has another "newFakeFlag" config flag in claim 2 that didn't exist in clam file 1.

```console
...
CNIs: Differences
FIELD                           CLAIM 1      CLAIM 2
/clus0-0/0/cniVersion           1.0.0        1.0.1
/clus0-1/0/plugins/0/ipMasq     true         false

CNIs: Only in CLAIM 1
<none>

CNIs: Only in CLAIM 2
/clus0-1/0/plugins/0/newFakeFlag=true
...
```

 Currently, the following sections are compared, in this order:

* claim.versions
* claim.Results
* claim.configurations.Config
* claim.nodes.cniPlugins
* claim.nodes.csiDriver
* claim.nodes.nodesHwInfo
* claim.nodes.nodeSummary

### How to build the tnf tool

The `tnf` tool is located in the repo's `cmd/tnf` folder. In order to compile it, just run:

```console
make build-tnf-tool
```

### Examples

#### Compare a claim file against itself: no differences expected

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="../assets/images/claim-compare-self.svg" width="100%" height=auto></object>
<!-- markdownlint-disable MD033 -->

#### Different test cases results

Let's assume we have two claim files, claim1.json and claim2.json, obtained from two CNF Certification Suite runs in the same cluster.

During the second run, there was a test case that failed. Let's simulate it modifying manually the second run's claim file to switch one test case's state from "passed" to "failed".

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="../assets/images/claim-compare-results.svg" width="100%" height=auto></object>
<!-- markdownlint-disable MD033 -->

#### Different cluster configurations

First, let's simulate that the second run took place in a cluster with a different OCP version. As we store the OCP version in the claim file (section claim.versions), we can also modify it manually.
The versions section comparison appears at the very beginning of the `tnf claim compare` output:

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="../assets/images/claim-compare-versions.svg" width="100%" height=auto></object>
<!-- markdownlint-disable MD033 -->

Now, let's simulate that the cluster was a bit different when the second CNF Certification Suite run was performed. First, let's make a manual change in claim2.json to emulate a different CNI version in the first node.

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="../assets/images/claim-compare-cni.svg" width="100%" height=auto></object>
<!-- markdownlint-disable MD033 -->

Finally, we'll simulate that, for some reason, the first node had one label removed when the second run was performed:

<!-- markdownlint-disable MD033 -->
<object type="image/svg+xml" data="../assets/images/claim-compare-nodes.svg" width="100%" height=auto></object>
<!-- markdownlint-disable MD033 -->
