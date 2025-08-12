CommandMock.ExecCommandContainer]
    B --> C{lock}
    C --> D[append to calls]
    D --> E[ExecCommandContainerFunc]
    E --> F{returns (stdout, stderr, err)}
    F --> G[unlock]
    G --> H[Test Assertions on calls]
```

This diagram visualizes the control flow from test code through the mock to the user‑supplied function and back.
