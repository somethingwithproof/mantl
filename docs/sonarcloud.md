# SonarCloud analysis

Mantl uses the `somethingwithproof` SonarCloud organization and project key
`somethingwithproof_mantl`. Mantl is currently public on GitHub and can use a
public SonarCloud project. The preflight check reads repository visibility from
the GitHub event. If the repository becomes private, it requires a private
SonarCloud project and an organization plan that supports it; source upload to a
public analysis project is refused.

Provision the project with `main` as its main branch and bind it to
`somethingwithproof/mantl` in SonarCloud. Disable automatic analysis when using
this CI-based workflow. Set `SONAR_TOKEN` as a GitHub Actions repository secret
using an analysis credential with access to this project. Read credentials from
Keychain or the environment at runtime; never commit or print their values.

The `SonarCloud quality gate` job runs on main pushes, pull requests and manual
dispatches. Missing authentication, an inaccessible project, coverage generation
failure, an incompatible project visibility, analysis failure or a rejected quality gate
fails the job before source upload where applicable. Fork pull
requests cannot access repository secrets; a maintainer must run the reviewed
change on a branch in this repository before it can pass the gate.

The scanner waits up to five minutes for the server's quality-gate result. Keep
the quality gate focused on new code and review the first main-branch analysis
before enabling merge enforcement. After a successful analysis, add the exact
`SonarCloud quality gate` check from GitHub Actions to main's required status
checks, preserving DCO and existing branch protection. A scanner failure or a
missing project must never be recorded as a passing quality result.

## Local coverage

Run through the runtimes pinned in `mise.toml`:

```sh
mise exec -- make go-coverage
mise exec -- python -m venv .venv-sonar
mise exec -- .venv-sonar/bin/python -m pip install -r requirements-test.txt PyYAML==6.0.3
mise exec -- .venv-sonar/bin/python -m pytest tests/unit/ --cov=scripts --cov-report=xml:coverage-python.xml
```

Go coverage uses the same first-party package list as `make go-test`, including
controllers and CLI packages while excluding vendored charts. Python coverage
measures the release and Sonar verification scripts; the unit suite's
infrastructure checks are not represented as application-code coverage.

Sonar classifies Go and Python test files as tests and excludes vendored charts,
provider caches, virtual environments, generated deepcopy code and build output.
Example applications remain in the analysis scope; their missing test coverage
is not hidden. The two coverage reports are uploaded as short-lived CI artifacts.
