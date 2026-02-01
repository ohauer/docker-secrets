## Description

<!-- Provide a clear and concise description of your changes -->

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring
- [ ] Test improvement
- [ ] CI/CD improvement
- [ ] Dependency update

## Related Issues

<!-- Link to related issues using #issue_number -->

Closes #
Relates to #

## Motivation and Context

<!-- Why is this change required? What problem does it solve? -->

## Changes Made

<!-- List the specific changes made in this PR -->

- 
- 
- 

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Test Configuration

- Go version:
- OS:
- Vault/OpenBao version (if applicable):

### Test Cases

- [ ] Unit tests pass (`make test`)
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Tested with Vault
- [ ] Tested with OpenBao
- [ ] Tested with TLS
- [ ] Tested in Docker container

### Test Evidence

<!-- Provide command output, screenshots, or logs demonstrating the tests -->

```bash
# Example test output
make test
```

## Documentation

<!-- Check all that apply -->

- [ ] Updated README.md
- [ ] Updated docs/ directory
- [ ] Updated code comments
- [ ] Updated examples
- [ ] Updated CHANGELOG.md (if exists)
- [ ] No documentation needed

## Checklist

<!-- Ensure all items are completed before requesting review -->

### Code Quality

- [ ] My code follows the project's style guidelines (`.kiro/steering/behavior/go_project_practices.md`)
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings
- [ ] I have run `make lint` and fixed all issues
- [ ] I have run `make fmt` to format the code

### Testing

- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have tested with race detector (`go test -race`)
- [ ] Test coverage has not decreased

### Documentation

- [ ] I have updated the documentation accordingly
- [ ] I have added examples for new features
- [ ] I have updated environment variable documentation (if applicable)

### Security

- [ ] My changes do not introduce security vulnerabilities
- [ ] I have not exposed secrets in logs or error messages
- [ ] I have followed secure coding practices
- [ ] I have considered the security implications of my changes

### Breaking Changes

<!-- If this is a breaking change, describe the impact and migration path -->

- [ ] This PR introduces breaking changes
- [ ] I have documented the breaking changes
- [ ] I have provided a migration guide

**Breaking Changes Description:**
<!-- Describe what breaks and how users should migrate -->

## Screenshots

<!-- If applicable, add screenshots to help explain your changes -->

## Performance Impact

<!-- Describe any performance implications of your changes -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance degraded (explain why this is acceptable)

**Performance Notes:**
<!-- Add benchmarks or profiling results if applicable -->

## Deployment Notes

<!-- Any special deployment considerations? -->

- [ ] Requires configuration changes
- [ ] Requires database migration
- [ ] Requires environment variable updates
- [ ] Requires dependency updates
- [ ] No special deployment steps

**Deployment Instructions:**
<!-- Provide step-by-step deployment instructions if needed -->

## Rollback Plan

<!-- How can this change be rolled back if issues are discovered? -->

## Additional Context

<!-- Add any other context about the PR here -->

## Reviewer Notes

<!-- Any specific areas you'd like reviewers to focus on? -->

---

## For Maintainers

<!-- Maintainers: Complete before merging -->

- [ ] PR title follows conventional commit format
- [ ] All CI checks pass
- [ ] Code review completed
- [ ] Documentation reviewed
- [ ] Security implications reviewed
- [ ] Breaking changes documented
- [ ] Release notes updated (if applicable)
