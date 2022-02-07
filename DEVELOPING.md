# Test Development Guide

Currently, tests are all CLI driven.  That means that the commands executed in test implementations must be made
available in the target container/machine/shell's `$PATH`.  Future work will address incorporating REST-based tests.
UI-driven tests are considered out of scope for `test-network-function`.

## General Test Writing Guidelines

In general, tests should adhere to the following principles:
* Tests should be platform independent when possible, and platform-aware when not.
* Tests must be runnable in a variety of contexts (i.e., `oc`, `ssh`, and `shell`).  Internally, we have developed a
variety of `interactive.Context` implementations for each of these.  In general, so long as your command does not depend
on specific prompts, the framework handles the context transparently.
* Tests must implement the `tnf.Tester` interface.
* Tests must implement the `reel.Handler` interface.
* Tests must be accompanied by appropriate unit tests.
* Tests adhere to the strict quality and style guidelines set forth in [CONTRIBUTING.md](CONTRIBUTING.md).

## Test Identifiers

Each `tnf.Tester` implementation *must* have a unique identifier.  In practice, `tnf.Tester` implementations are the
building blocks of larger test suites, and each implementation ought to have a means of identification.

An [`identifier.Identifier`](pkg/tnf/identifier/identifier.go) is the mechanism used to hold this meta information.
Please see the implementation for details.  Essentially, an Identifier is just a URL and a Semantic Version.

### Example of creating an Identifier

To create an identifier for your test, go to  [`identifiers.go`](pkg/tnf/identifier/identifiers.go).  Create a constant
for the URL, and add the `TestCatalogEntry` to the `Catalog` map such as:

```go
listRootDirectoryFilesURL = "http://test-network-function.com/tests/listRootDirectoryFiles"

...
var Catalog = map[string]TestCatalogEntry{
...
    listRootDirectoryFilesURL: {
        URL: listRootDirectoryFilesURL,
        Description: "A test to list the files at the root of the file system.",
        Type: Normative,
    }
...
}
```

Reference the exported URL constant in your `tnf.Handler` `GetIdentifier()` implementation.

*Note*: JSON tests should also involve creation of an identifier using the same Go-based methodology for `1.0`.

### Identifier Re-use

Identifiers can be reused, but they should follow the rules of [semantic versioning](https://semver.org/).  Namely, the
following versioning should be utilized:

Version Level|Description
---|---
Major|API incompatible changes.
Minor|Add functionality that is backwards compatible.
Patch|Backwards compatible bug fixes.

*Note*: If the premise of the test changes drastically, consider creating a new identifier instead of bumping the major
version of an existing one.

## Language options for writing test implementations

There are two options for writing test implementations:
1) JSON
2) Go

The JSON approach is significantly quicker to implement, and should be preferred when possible.

## Writing a simple CLI-oriented test in JSON

Most tests just involve sending commands and validating output within a single shell context.  For example, open an
interactive shell to a container and ping a target host.  On the command line, this might be done similar to the
following:

```shell-script
oc exec -it <podName> -c <containerName> -- sh
ping -c <count> <destination>
```

We would expect the ping command to take approximately 5 seconds, as most implementations of `ping` default to 1 second
for inter-packet gap.  After the command completes, we would expect a summary to be output.  Thus, the whole
interaction would be similar to the following:

```shell-script
% oc exec -it test -c test -- sh
sh-4.2# ping -c 5 www.redhat.com
PING e3396.dscx.akamaiedge.net (23.34.95.235) 56(84) bytes of data.
64 bytes from a23-34-95-235.deploy.static.akamaitechnologies.com (23.34.95.235): icmp_seq=1 ttl=61 time=16.1 ms
64 bytes from a23-34-95-235.deploy.static.akamaitechnologies.com (23.34.95.235): icmp_seq=2 ttl=61 time=22.8 ms
64 bytes from a23-34-95-235.deploy.static.akamaitechnologies.com (23.34.95.235): icmp_seq=3 ttl=61 time=24.5 ms
64 bytes from a23-34-95-235.deploy.static.akamaitechnologies.com (23.34.95.235): icmp_seq=4 ttl=61 time=23.6 ms
64 bytes from a23-34-95-235.deploy.static.akamaitechnologies.com (23.34.95.235): icmp_seq=5 ttl=61 time=18.3 ms

--- e3396.dscx.akamaiedge.net ping statistics ---
5 packets transmitted, 5 received, 0% packet loss, time 4007ms
rtt min/avg/max/mdev = 16.163/21.128/24.579/3.276 ms
sh-4.2#
```

We have now established the needed commands for a basic test.  We have established that:
1) The test should use an `oc` based context to connect to pod `test` container `test` interactively.
2) The test should issue a `ping` command with a count of `5` against `www.redhat.com`.
3) After the `ping` command completes, we should inspect the summary to ensure that we received the expected number of
packets.

### Writing the test

Generic tests *must* abide by the [generic-test.schema.json](schemas/generic-test.schema.json) JSON Schema.  Let's
consider the simple [`ping` example](examples/ping.json).  It describes a test that pings "www.redhat.com" 5 times,
and gives a `tnf.SUCCESS` only if all 5 pings receive a response.

Let's walk through [`ping.json`](examples/ping.json) one key one at a time.

#### description

A human-readable description is required for every test.  This documentation proves invaluable for later reuse.

#### testResult

`testResult` is initialized as 0.  As of now, `tnf.Tester` mandates `Result()` returns an `int` value.  `0` corresponds
to `tnf.ERROR`.  In fact, all tests should start report `tnf.ERROR` as good practice, and only progress to
`tnf.SUCCESS` or `tnf.FAILURE` after inspecting a given match.

#### testTimeout

```testTimeout``` is the timeout for the test in nanoseconds.  For this example, we chose a duration of `10s` to perform
the ping test.

#### reelFirstStep

```reelFirstStep``` is the first `reel.Step` in the `REEL` finite state machine (FSM).  A `reel.Step` contains a
mandatory `timeout` and optional `execute` and `expect`.  *Note*: Remember to append `\n` to commands specified in
`execute`.  In this case, `execute` is exactly what you might suspect;  the ping command to www.redhat.com.

The `expect` field deserves further explanation.  `expect` is an array of regular expressions that might match from
issuing a `ping` command.  In this case, only one regular expression is expected, the ping summary.  However, if more
than one `expect` element exists, the matches are determined *in-order* and the first match found will complete the step.

Development of regular expressions falls outside of this tutorial.  [regex101.com](https://regex101.com/) provides a
useful interface to help design regular expressions for Go.

#### resultContexts

`resultContexts` is self describing;  if a regular expression in `reelFirstStep.expect` is matched, then we will want to
provide further logic to progressively determine whether the test is a `tnf.SUCCESS` or `tnf.FAILURE`.  Importantly,
these concepts are dependent on your own business logic.  In this example, we require a response to every ICMP request
packet sent.  A different implementation may be more lenient, and require only `numRequests - 1` ICMP responses.  The
implementation is completely up to the test writer.

Each `reelFirst.expect` must have a `resultContext`.  In this case, we only have one pattern, which is represented by
the ICMP summary regular expression:

```json
{
  "pattern": "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$",
  "defaultResult": 1,
  "composedAssertions": [
    {
      "assertions": [
        ...
      ],
      "logic": {
        "type": "and"
      }
    }
  ]
}
```

`defaultResult` is the test result returned if no `composedAssertion`s exist.  For example, if we omitted
`composedAssertions` above, then the mere fact that the ping summary matched would result in `tnf.SUCCESS`.  However,
since `composedAssertions` is provided, we must do further inspection.

In this case, only one `composedAssertion` is provided.  In case multiple `composedAssertion` are provided, each `composedAssertion` instances must evaluate
as `true`. Here is the single `composedAssertion` that we make in the test:

```json
{
  "assertions": [
    {
      "groupIdx": 1,
      "condition": {
        "type": "intComparison",
        "input": 5,
        "comparison": "=="
      }
    },
    {
      "groupIdx": 2,
      "condition": {
        "type": "intComparison",
        "input": 5,
        "comparison": "=="
      }
    }
  ],
  "logic": {
    "type": "and"
  }
}
```

Essentially, a `composedAssertion` allows us to make several sub-assertions using `logic`.  In this case, `and` is used,
meaning each `assertion` must evaluate as true.

There are two assertions in this example.  The first assertion is as follows:

```json
{
  "groupIdx": 1,
  "condition": {
    "type": "intComparison",
    "input": 5,
    "comparison": "=="
  }
}
```

This means that group 1 of the regular expression match for the ping summary pattern must equal `5`.  In this case,
group 1 is the number of ICMP requests transmitted.

The second assertion is as follows:

```json
{
  "groupIdx": 2,
  "condition": {
    "type": "intComparison",
    "input": 5,
    "comparison": "=="
  }
}
```

This means that group 2 of the regular expression match for the ping summary pattern must equal `5`.  IN this case,
group 2 is the number of ICMP requests received.

The whole of the JSON `pattern` and `composedAssertions` means the test will pass ONLY if:

* exactly 5 pings were sent
* exactly 5 responses were received

### Running your JSON test

Now that you have a sample JSON test defined, you can go ahead and run your JSON test in your development environment.
In order to run the test, you must first make the jsontest CLI.  Issue the following command:

```shell-script
./tnf jsontest shell examples/ping.json
```

You will get something similar to the following:

```shell-script
% ./tnf jsontest shell examples/ping.json
INFO[0000] Running examples/ping.json from a local shell context
2020/12/06 13:32:53 Sent: "ping -c 5 www.redhat.com\n"
2020/12/06 13:32:57 Match for RE: "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$" found: ["5 packets transmitted, 5 packets received, 0.0% packet loss" "5" "5" " packets" ""] Buffer: "PING e3396.dscx.akamaiedge.net (23.34.95.235): 56 data bytes\n64 bytes from 23.34.95.235: icmp_seq=0 ttl=59 time=17.661 ms\n64 bytes from 23.34.95.235: icmp_seq=1 ttl=59 time=25.993 ms\n64 bytes from 23.34.95.235: icmp_seq=2 ttl=59 time=26.353 ms\n64 bytes from 23.34.95.235: icmp_seq=3 ttl=59 time=25.725 ms\n64 bytes from 23.34.95.235: icmp_seq=4 ttl=59 time=22.403 ms\n\n--- e3396.dscx.akamaiedge.net ping statistics ---\n5 packets transmitted, 5 packets received, 0.0% packet loss\nround-trip min/avg/max/stddev = 17.661/23.627/26.353/3.302 ms\n"
INFO[0004] Test Result: 1
INFO[0004] Test Payload:
INFO[0004] {
    "description": "Pings www.redhat.com 5 times using the Unix ping executable.",
    "matches": [
        {
            "pattern": "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$",
            "before": "PING e3396.dscx.akamaiedge.net (23.34.95.235): 56 data bytes\n64 bytes from 23.34.95.235: icmp_seq=0 ttl=59 time=17.661 ms\n64 bytes from 23.34.95.235: icmp_seq=1 ttl=59 time=25.993 ms\n64 bytes from 23.34.95.235: icmp_seq=2 ttl=59 time=26.353 ms\n64 bytes from 23.34.95.235: icmp_seq=3 ttl=59 time=25.725 ms\n64 bytes from 23.34.95.235: icmp_seq=4 ttl=59 time=22.403 ms\n\n--- e3396.dscx.akamaiedge.net ping statistics ---",
            "match": "5 packets transmitted, 5 packets received, 0.0% packet loss"
        }
    ],
    "reelFirstStep": {
        "execute": "ping -c 5 www.redhat.com\n",
        "expect": [
            "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$"
        ],
        "timeout": 10000000000
    },
    "resultContexts": [
        {
            "pattern": "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$",
            "composedAssertions": [
                {
                    "assertions": [
                        {
                            "groupIdx": 1,
                            "condition": {
                                "type": "intComparison",
                                "input": 5,
                                "comparison": "=="
                            }
                        },
                        {
                            "groupIdx": 2,
                            "condition": {
                                "type": "intComparison",
                                "input": 5,
                                "comparison": "=="
                            }
                        }
                    ],
                    "logic": {
                        "type": "and"
                    }
                }
            ],
            "defaultResult": 1
        }
    ],
    "testResult": 1,
    "testTimeout": 10000000000
}
```

Note that `testResult` is 1, indicating `tnf.SUCCESS`.

If you wish to explore the `oc` and `ssh` variants of `jsontest-cli`, please consult the following:

```shell-script
./jsontest -h
```

### Including a JSON-based test in a Ginkgo Test Suite

See the [diagnostic](test-network-function/diagnostic/suite.go) test suite for an example of this.

### Including templated JSON-based tests

Often times, tests require arguments.  For example, if you were to write a test which involves testing `ping` to a
particular destination, perhaps derived dynamically, the destination would need to be configurable.  In this case,
Go templates can be used to render a JSON-based test.  [ping.json.tpl](./examples/generic/template/ping.json.tpl) is an
example of a JSON-test which contains a `HOST` argument, and
[ping.values.yaml](./examples/generic/template/ping.values.yaml) provides the necessary values.  Tests can be rendered
using something similar to:

```go
templateFile := path.Join("examples", "generic", "template", "ping.json.tpl")
schemaPath := path.Join("schemas", "generic-test.schema.json")
valuesFile := path.Join("examples", "generic", "template", "ping.values.yaml")
tester, handlers, result, err := generic.NewGenericFromTemplate(templateFile, schemaPath, valuesFile)
```

## Writing a simple CLI-oriented test in Go

A `test-network-function` test must implement `tnf.Tester` and `reel.Handler` Go `interface`s.  The `tnf.Tester`
interface defines the contract required for a CLI-based test, and `reel.Handler` defines the Finite State Machine (FSM)
contract for executing the test.  A basic example is [ping.go](pkg/tnf/handlers/ping/ping.go).

We will go through implementing the required interfaces one at a time below:

### Implementing `ping.go` `tnf.Tester`

For a test to implement the `tnf.Tester` interface, it must provide definitions for `Args`, `Timeout` and `Result`.
These are the set of accessor methods used to define characteristics of the test, as well as the actual result of the
test.  Note, this does not include any expected results;  those need to be defined later.

First create a type called `Ping` which is capable of storing `result`, `timeout` and `args` variables.  Additionally,
we will restrict our test by mandating that a positive integer `count` and `destination` string must be provided during
the time of instantiation.

```go
// Ping provides a ping test implemented using command line tool `ping`.
type Ping struct {
    result      int
    timeout     time.Duration
    args        []string
    count       int
    destination string
}
```

To better enforce data encapsulation, please only export (capitalize) variables that are absolutely needed.  For
example, we use `count` not `Count` above.

If you look at the [`tnf.Tester`](pkg/tnf/test.go) interface definition, you will notice that the data types for
`result`, `timeout` and `args` match the return types for the mandated functions.

```go
type Tester interface {
	Args() []string
	Timeout() time.Duration
	Result() int
}
```

After creating the struct, define the accessors similar to the following:

```go
// Args returns the command line args for the test.
func (p *Ping) Args() []string {
	return []string{"ping", "-c", p.count, p.destination}
}

// Timeout returns the timeout in seconds for the test.
func (p *Ping) Timeout() time.Duration {
	return p.timeout
}

// Result returns the test result.
func (p *Ping) Result() int {
	return p.result
}
```

The `Args()` implementation deserves some explaining.  `Args()` is an array of the commands line argument strings.  In
this example, the command `ping -c 5 www.redhat.com` is represented as `string[]{"ping", "-c", "5", "www.redhat.com"}`.
In other words, the elements are all of the white-space separated string components of the command.

That completes our `ping.go` `tnf.Tester` implementation!  Next, implement the logic of the `reel.Handler` FSM.

### Implementing `ping.go` `reel.Handler`

The easy part is out of the way.  Implementing `reel.Handler` is slightly more involved, but should make sense after
completing this part of the tutorial.  [reel.go](pkg/tnf/reel/reel.go) defines the `reel.Handler` interface:

```go
// A Handler implements desired programmatic control.
type Handler interface {
	// ReelFirst returns the first step to perform.
	ReelFirst() *Step

	// ReelMatch informs of a match event, returning the next step to perform.  ReelMatch takes three arguments:
	// `pattern` represents the regular expression pattern which was matched.
	// `before` contains all output preceding `match`.
	// `match` is the text matched by `pattern`.
	ReelMatch(pattern string, before string, match string) *Step

	// ReelTimeout informs of a timeout event, returning the next step to perform.
	ReelTimeout() *Step

	// ReelEOF informs of the eof event.
	ReelEOF()
}
```

We will handle describing implementing each of these methods one by one.

#### Implementing `ping.go` `ReelEOF()`

`ReelEOF` is used to define the callback executed when EOF is encountered in the context.  Unexpected interruptions to
`ssh` or `oc` session are common reasons for EOF.

For the case of ping we can make this simple.  Since we require a `count` for `ping`, we don't need to do anything
particular for EOF.

```go
// ReelEOF does nothing;  ping requires no intervention on EOF.
func (p *Ping) ReelEOF() {
}
```

#### Implementing `ping.go` `ReelTimeout()`

`ReelTimeout` is used to define the callback executed when a test times out.

When a ping test times out, we probably ought to issue a `CTRL+C` in order to exit early and prepare the context for
future commands.

```go
// ReelTimeout returns a step which kills the ping test by sending it ^C.
func (p *Ping) ReelTimeout() *reel.Step {
	return &reel.Step{Execute: "\003"}
}
```

#### Implementing `ping.go` `ReelFirst()`

Since we supply `tnf.Test` `Args()`, we do not need to include anything for `Execute` in the returned `reel.Step`.

```go
// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (p *Ping) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{`(?m)connect: Invalid argument$`, `(?m)(\d+) packets transmitted, (\d+)( packets){0,1} received, (?:\+(\d+) errors)?.*$`},
		Timeout: p.timeout,
	}
}
```

Note:  The ordering of `Expect` matters!  The framework matches `Expect` elements in index-ascending order.

### Implementing `ping.go` `ReelMatch()`

This is likely the hardest part of any test implementation.  `ReelMatch` needs to decipher what is matched, and assign
the appropriate result to the `tnf.Test`.  Let's take a look at the implementation provided for
[ping.go](pkg/tnf/handlers/ping/ping.go):

```go
// ReelMatch parses the ping statistics and set the test result on match.
// The result is success if at least one response was received and the number of
// responses received is at most one less than the number received (the "missing"
// response may be in flight).
// The result is error if ping reported a protocol error (e.g. destination host
// unreachable), no requests were sent or there was some test execution error.
// Otherwise the result is failure.
// Returns no step; the test is complete.
func (p *Ping) ReelMatch(_ string, _ string, match string) *reel.Step {
	re := regexp.MustCompile(`(?m)connect: Invalid argument$`)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		p.result = tnf.ERROR
	}
	re = regexp.MustCompile(SuccessfulOutputRegex)
	matched = re.FindStringSubmatch(match)
	if matched != nil {
		// Ignore errors in converting matches to decimal integers.
		// Regular expression `stat` is required to underwrite this assumption.
		p.transmitted, _ = strconv.Atoi(matched[1])
		p.received, _ = strconv.Atoi(matched[2])
		p.errors, _ = strconv.Atoi(matched[4])
		switch {
		case p.transmitted == 0 || p.errors > 0:
			p.result = tnf.ERROR
		case p.received > 0 && (p.transmitted-p.received) <= 1:
			p.result = tnf.SUCCESS
		default:
			p.result = tnf.FAILURE
		}
	}
	return nil
}
```

Essentially, since `ReelMatch()` always returns `nil` this function is the final state for the `reel.Step` FSM.  For
more advanced tests, `ReelMatch()` can be called an arbitrary number of times.  In this example, `ReelMatch()` is only
called once.

The logic for determining the test result is up to the test writer.  This particular implementation analyzes the match
output to determine the result.
1) If the provided `destination` results in an `Invalid Argument`, then `tnf.ERROR` is returned.
2) If the ping summary regular expression matched, then:
* `tnf.ERROR` if there were PING transmit errors
* `tnf.SUCCESS` if a maximum of a single packet was lost
* `tnf.FAILURE` for any other case.

### Including `ping.go` in a Ginkgo Test Suite

An example of using `ping.go` from within a Ginkgo test spec is included in
[suite.go](test-network-function/generic/suite.go)'s `testPing` method.  Roughly, the code should resemble the
following:

```go
// 1. Create the Test.
pingTester := ping.NewPing(defaultTimeout, targetPodIPAddress, count)
test, err := tnf.NewTest(oc.GetExpecter(), pingTester, []reel.Handler{pingTester}, oc.GetErrorChannel())
gomega.Expect(err).To(gomega.BeNil())

// 2. Run the Test.
testResult, err := test.Run()
gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
gomega.Expect(err).To(gomega.BeNil())

// 3. Inspect the Results.
transmitted, received, errors := pingTester.GetStats()
gomega.Expect(received).To(gomega.Equal(transmitted))
gomega.Expect(errors).To(gomega.BeZero())
```

## Writing `ping.go` test Summary

You should now have the appropriate knowledge to write your own test implementation.  There are a variety of
implementations included out of the box in the [handlers](pkg/tnf/handlers) directory.

This guide does not cover unit testing the Test, nor does it cover managing test-specific configuration.  Please see the
examples of existing tests in the codebase for how to do these things.

## Writing custom PTY interactive.Context Implementations

Although `test-network-function` includes built in `interactive.Context` implementations for `oc`, `shell` and `ssh`,
there are many cases in which you may need a completely new PTY.  For example, networking software (including vpp, Cisco
IOS, etc.) often includes interactive PTY-based menus.  In such a case, you will need a new `interactive.Context` to
communicate with the underlying Shell.

In such cases, consider using `interactive.SpawnGenericPTYFromYAMLFile(...)` or its corollary
`interactive.SpawnGenericPTYFromYAMLTemplate(...)` which can be templated using Go `text/template` language.  Examples
of such PTY implementations can be found in [examples/pty](./examples/pty).

## Processing CLI Output: A note about `oc` and `jq`

The current tests frequently use `jq` to process structured output from `oc -o json`. `oc` also allows use of
[Go Templates](https://www.openshift.com/blog/customizing-oc-output-with-go-templates) for processing structured output.
This is potentially more powerful than using `jq` as it allows building highly customized output of multiple resources
simultaneously without adding dependencies. Conversely `jq` is widely available and commonly used, and has been
sufficient for all cases so far. It is up to the author of a contribution to decide which approach is best suited to the
task at hand.

While the likely use of Go Templates is at the complex end of the spectrum, as a simple example the command used in `CONTAINER_COUNT` to find the number of containers in a pod is currently using `jq`:

```shell-script
oc get pod %s -n %s -o json | jq -r '.spec.containers | length'
```

The same result could be achieved using a Go Template:

```shell-script
oc get pod %s -n %s -o go-template='{{len .spec.containers}}{{"\n"}}'
```

## Adding new handler

To facilitate adding new handlers, the "tnf" utility has been created to help developers to avoid writing repetitive code. The tnf tool [source code is here](cmd/tnf) and can be built with the following command:
```shell-script
make build-tnf-tool
```

To generate a new handler named MyHandler, use the options "generate handler" as in the next example:
```shell-script
./tnf generate handler MyHandler
```

The generated code has a template and creates the necessary headers.
The result is folder "myhandler" located in /pkg/tnf/handlers/myhandler that includes 3 files by handler template.
The command relays on golang templates located in [pkg/tnf/handlers/handler_template](pkg/tnf/handlers/handler_template), so in case the "tnf" utility is executed outside the test-network-function root folder, the user can export the environment variable TNF_HANDLERS_SRC pointing to an existing "handlers" relative/absolute folder path.
```shell-script
 export TNF_HANDLERS_SRC=other/path/pkg/tnf/handlers
```

## Adding information to claim file

The result of each test execution is included in the claim file.
Sometimes it is convenient to add informational messages regarding the test execution.
In order to add informational messages to your test use the function `ginkgo.GinkgoWriter`.
This function adds an additional message that will appear in the `CapturedTestOutput` section of the claim file, together with the output of the by directives.
Each added message will be written to claim file even if test failed or error occurred in the middle of the test.

Example usage:
```go
ginkgo.It("Should do what I tell it to do", func(){
  // do some more work
  // add info
  _, err := ginkgo.GinkgoWriter.Write([]byte("important info part 1"))
  if err != nil {
    log.Errorf("Ginkgo writer could not write because: %s", err)
  }

  // more work
  // more info
  _, err := ginkgo.GinkgoWriter.Write([]byte("important info part 2"))
  if err != nil {
    log.Errorf("Ginkgo writer could not write because: %s", err)
  }
  // error
  if err != nil {
    return
  }
  // last info
  _, err := ginkgo.GinkgoWriter.Write([]byte("important info part last"))
  if err != nil {
    log.Errorf("Ginkgo writer could not write because: %s", err)
  }
})
```