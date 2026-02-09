# Fuzzing

This project includes fuzz tests to protect public interfaces from invalid input.

## What is Fuzzing?

Fuzzing is an automated testing technique that provides random, malformed, or unexpected input to find bugs, crashes, and security vulnerabilities. Go has built-in fuzzing support since Go 1.18.

## Fuzz Tests

### Config Loading (`internal/config/loader_fuzz_test.go`)
- **FuzzConfigLoad**: Tests YAML parsing with malformed input
- **FuzzValidateFilePath**: Tests path validation with traversal attacks

**Attack vectors tested**:
- Malformed YAML syntax
- Deeply nested structures
- Extreme values
- Path traversal attempts (`../../../etc/passwd`)
- Null bytes and special characters

### File Path Validation (`internal/filewriter/writer_fuzz_test.go`)
- **FuzzValidatePath**: Tests path validation logic
- **FuzzValidateMode**: Tests file mode validation

**Attack vectors tested**:
- Path traversal (`../`, `..`)
- Relative paths
- Null bytes (`\x00`)
- Newlines and control characters
- Unicode attacks (right-to-left override)
- Windows paths on Linux
- World-writable modes
- Invalid octal values

### Template Rendering (`internal/template/engine_fuzz_test.go`)
- **FuzzRender**: Tests template engine with malicious templates

**Attack vectors tested**:
- Malformed template syntax
- Infinite loops
- Large templates (DoS)
- Non-existent fields
- Nested pipes

## Running Fuzz Tests

### Quick Test (10s per test)
```bash
make fuzz
```

### Extended Fuzzing (1 minute per test)
```bash
go test -fuzz=FuzzValidatePath -fuzztime=1m ./internal/filewriter
go test -fuzz=FuzzValidateMode -fuzztime=1m ./internal/filewriter
go test -fuzz=FuzzRender -fuzztime=1m ./internal/template
go test -fuzz=FuzzConfigLoad -fuzztime=1m ./internal/config
go test -fuzz=FuzzValidateFilePath -fuzztime=1m ./internal/config
```

### Continuous Fuzzing (until failure)
```bash
go test -fuzz=FuzzConfigLoad ./internal/config
```

Press `Ctrl+C` to stop.

## Interpreting Results

### Success
```
fuzz: elapsed: 5s, execs: 1331763 (258675/sec), new interesting: 1 (total: 16)
PASS
```

### Failure
If a fuzz test finds a crash, it will:
1. Save the failing input to `testdata/fuzz/FuzzTestName/`
2. Show the panic/error
3. Fail the test

Example:
```
--- FAIL: FuzzValidatePath (0.50s)
    --- FAIL: FuzzValidatePath (0.00s)
        testing.go:1349: panic: runtime error: index out of range
```

The failing input is saved and can be replayed:
```bash
go test -run=FuzzValidatePath/crash-hash ./internal/filewriter
```

## Corpus Management

Fuzz tests build a corpus of interesting inputs in `testdata/fuzz/`:
```
testdata/fuzz/
├── FuzzConfigLoad/
│   ├── seed1
│   └── seed2
├── FuzzValidatePath/
│   └── seed1
└── FuzzRender/
    └── seed1
```

These are automatically used in future runs to maintain coverage.

## CI Integration

Add to CI pipeline:
```yaml
- name: Fuzz tests
  run: make fuzz
```

For longer fuzzing in CI:
```yaml
- name: Extended fuzz tests
  run: |
    go test -fuzz=. -fuzztime=5m ./internal/filewriter
    go test -fuzz=. -fuzztime=5m ./internal/template
    go test -fuzz=. -fuzztime=5m ./internal/config
```

## Security Benefits

Fuzzing protects against:
- **Path Traversal**: Prevents reading/writing outside allowed directories
- **Template Injection**: Prevents malicious template execution
- **DoS Attacks**: Catches infinite loops and resource exhaustion
- **Input Validation Bypass**: Finds edge cases in validation logic
- **Panic/Crash Bugs**: Discovers unexpected input handling

## Best Practices

1. **Seed with known attack vectors**: Include common exploits in `f.Add()`
2. **Run regularly**: Include in CI/CD pipeline
3. **Extended runs**: Periodically run longer fuzzing sessions
4. **Review corpus**: Check `testdata/fuzz/` for interesting cases
5. **Fix and re-test**: When bugs are found, fix and verify with saved input

## References

- [Go Fuzzing Documentation](https://go.dev/doc/fuzz/)
- [OWASP Path Traversal](https://owasp.org/www-community/attacks/Path_Traversal)
- [Template Injection](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/18-Testing_for_Server-side_Template_Injection)
