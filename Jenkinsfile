// GitHub Actions owns required checks, Sonar analysis and release publication.
// This compatibility entrypoint runs local checks on an existing Jenkins worker.
node {
    timeout(time: 30, unit: 'MINUTES') {
        stage('Checkout') {
            checkout scm
        }
        stage('Pinned toolchain') {
            sh '''
                set -eu
                mise trust --yes
                mise install go python terraform kubectl kustomize uv golangci-lint
            '''
        }
        stage('Go unit tests and lint') {
            sh '''
                set -eu
                mise exec -- make go-test
                mise exec -- golangci-lint run
            '''
        }
        stage('Python contracts and dependency locks') {
            sh '''
                set -eu
                mise exec -- python -m ci.lock_python_dependencies --check
                mise exec -- python -m venv .venv-jenkins
                mise exec -- .venv-jenkins/bin/python -m pip install \
                    --only-binary :all: --require-hashes -r requirements-test.txt
                mise exec -- .venv-jenkins/bin/python -m pytest tests/unit/
            '''
        }
        stage('Terraform formatting') {
            sh '''
                set -eu
                mise exec -- terraform fmt -check -recursive infra/terraform/blueprints
                mise exec -- terraform fmt -check -recursive infra/terraform/modules
            '''
        }
    }
}
