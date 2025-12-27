Pipelines

- sample-pipeline.json: Deploys inline manifests to apps-dev (no Bake).
- sample-pipeline-bake.json: Bakes with Kustomize from this repo and deploys to apps-dev using a GitRepo artifact.

Before importing sample-pipeline-bake.json:
1) Ensure your Spinnaker overlay enables artifacts.github and artifacts.gitRepo and that a token is provided via github.env.
2) Edit the JSON to set:
   - reference: https://github.com/YOUR_ORG/YOUR_REPO.git
   - version: main (or another branch)
3) Confirm the account names:
   - artifact account: github-repo
   - deploy account: default-account
4) The Kustomize path is examples/spinnaker/apps/sample-app; change if needed.
