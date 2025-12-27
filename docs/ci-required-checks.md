# Required checks (GitHub)

We recommend marking the following workflows as required in your branch protection rules (contexts include the workflow name "CI"):
- CI / policy-tests (strict on main/tags)
- CI / platform-health
- CI / kyverno-validate
- CI / kind-smoke
- CI / e2e-hello
- CI / gateway-smoke
- CI / external-dns-rfc2136-e2e

Using GitHub UI:
1. Settings -> Branches -> Branch protection rules -> Edit main
2. Enable "Require status checks to pass before merging" and select the workflows above.

Using GitHub CLI (script):
```
# Requires gh auth and admin rights
scripts/set-required-checks.sh main
```
