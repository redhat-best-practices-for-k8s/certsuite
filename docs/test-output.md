# Test Output

## Claim File

The test suite generates an output file, named **claim** file. This file is considered as the proof of CNFs test run, evaluated by Red Hat when **certified** status is considered.

This file describes the following

* The system(s) under test
* The tests that are executed
* The outcome of the executed / skipped tests

**How to add a CNF platform test result to the existing claim file?**

```go
go run cmd/tools/cmd/main.go claim-add --claimfile=claim.json
--reportdir=/home/$USER/reports
```
 **Args**:  
`
--claimfile is an existing claim.json file`

`
--repordir :path to test results that you want to include.
`

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
