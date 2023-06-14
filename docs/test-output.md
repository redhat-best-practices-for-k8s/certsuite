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

* [schema](https://github.com/test-network-function/test-network-function-claim/blob/main/schemas/claim.schema.json).  
* [Guide](https://redhat-connect.gitbook.io/openshift-badges/badges/cloud-native-network-functions-cnf).

## Execution logs

The test suite also saves a copy of the execution logs at [test output directory]/tnf-execution.log

## Results artifacts zip file

After running all the test cases, a compressed file will be created with all the results files and web artifacts to review them. The file has a UTC date-time prefix and looks like this:

20230620-110654-cnf-test-results.tar.gz

The "20230620-110654" sample prefix means "June-20th 2023, 11:06:54"

This is the content of the tar.gz file:
- claim.json
- cnf-certification-tests_junit.xml
- claimjson.js
- classification.js
- results-embed.html
- results.html

This file serves two different purposes:
1. Make it easier to store and send the test results for review.
2. View the results in the html web page. In addition, the web page (either results-embed.thml or results.html) has a selector for workload type and allows the parter to introduce feedback for each of the failing test cases for later review from Red Hat. It's important to note that this web page needs the `claimjson.js` and `classification.js` files to be in the same folder as the html files to work properly.
