BaseImageInfo.runCommand --> ExecCommandContainer;
    BaseImageInfo.runCommand --> Error("exec failed");
    BaseImageInfo.runCommand --> New("non-zero exit");
```

This diagram shows that `runCommand` delegates to the container executor and then wraps any failure in a descriptive error.

---

### Summary  

* **Where**: `isredhat/isredhat.go`, line 70.  
* **What it does**: Executes a shell command inside the test container and returns its stdout or an informative error.  
* **Why it matters**: Enables the test suite to introspect base images for Red‑Hat markers without exposing command execution logic elsewhere.
