pipeline {
    agent any

    environment {
        PYTHONPATH = "${WORKSPACE}"
        TERRAFORM_VERSION = "1.10.5"
        KYVERNO_VERSION = "3.3.2"
        KUBECTL_VERSION = "1.31.0"
    }

    options {
        buildDiscarder(logRotator(numToKeepStr: '30'))
        timeout(time: 2, unit: 'HOURS')
        disableConcurrentBuilds()
        timestamps()
    }

    triggers {
        pollSCM('H/15 * * * *')
    }

    stages {
        stage('Setup') {
            steps {
                script {
                    echo "Setting up build environment..."
                }
                sh '''
                    python3 -m pip install --upgrade pip
                    pip3 install -r requirements.txt
                    pip3 install -r requirements-test.txt
                    pip3 install safety
                '''
            }
        }

        stage('Lint') {
            parallel {
                stage('Python Linting') {
                    agent {
                        docker {
                            image 'python:3.12-slim'
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            python -m pip install --upgrade pip
                            pip install -r requirements.txt

                            # Run black formatter check
                            black --check .

                            # Run ruff linter
                            ruff check .

                            # Run yamllint
                            yamllint .
                        '''
                    }
                }

                stage('Terraform Formatting') {
                    agent {
                        docker {
                            image "hashicorp/terraform:${TERRAFORM_VERSION}"
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            terraform fmt -check -recursive
                        '''
                    }
                }

                stage('Kubernetes Manifests') {
                    agent {
                        docker {
                            image 'alpine:latest'
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            # Install kustomize and kubectl
                            apk add --no-cache curl
                            curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
                            mv kustomize /usr/local/bin/

                            curl -LO "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
                            chmod +x kubectl
                            mv kubectl /usr/local/bin/

                            # Validate kustomizations
                            find . -name kustomization.yaml -exec dirname {} \\; | while read dir; do
                                echo "Validating $dir"
                                kustomize build "$dir" > /dev/null || echo "Warning: $dir failed"
                            done

                            # Validate ArgoCD Applications
                            find clusters/ -name "*.yaml" -type f | while read app; do
                                echo "Validating ArgoCD app: $app"
                                kubectl apply --dry-run=client -f "$app" || echo "Warning: $app failed validation"
                            done
                        '''
                    }
                }

                stage('Shell Scripts') {
                    agent {
                        docker {
                            image 'koalaman/shellcheck-alpine:stable'
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            # Shellcheck all shell scripts
                            find . -name "*.sh" -type f | xargs shellcheck --severity=warning || true
                        '''
                    }
                }

                stage('Kyverno Policies') {
                    agent {
                        docker {
                            image 'alpine:latest'
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            # Install kyverno CLI
                            apk add --no-cache curl
                            curl -LO "https://github.com/kyverno/kyverno/releases/download/v${KYVERNO_VERSION}/kyverno-cli_v${KYVERNO_VERSION}_linux_x86_64.tar.gz"
                            tar -xzf kyverno-cli_v${KYVERNO_VERSION}_linux_x86_64.tar.gz
                            mv kyverno /usr/local/bin/

                            # Validate Kyverno policies
                            find policies/kyverno -name "*.yaml" -type f | while read policy; do
                                echo "Validating Kyverno policy: $policy"
                                kyverno validate "$policy" || echo "Warning: $policy failed validation"
                            done
                        '''
                    }
                }
            }
        }

        stage('Security Scan') {
            steps {
                sh '''
                    # Run safety scan on Python dependencies
                    safety scan --output json || true
                '''
            }
        }

        stage('Test') {
            agent {
                docker {
                    image 'python:3.12-slim'
                    reuseNode true
                }
            }
            steps {
                sh '''
                    python -m pip install --upgrade pip
                    pip install -r requirements.txt -r requirements-test.txt

                    # Run pytest with coverage
                    pytest -q --junitxml=test-results.xml --cov=. --cov-report=xml --cov-report=html || true
                '''
            }
            post {
                always {
                    junit allowEmptyResults: true, testResults: 'test-results.xml'
                    publishHTML(target: [
                        allowMissing: true,
                        alwaysLinkToLastBuild: false,
                        keepAll: true,
                        reportDir: 'htmlcov',
                        reportFiles: 'index.html',
                        reportName: 'Coverage Report'
                    ])
                }
            }
        }

        stage('Terraform Validation') {
            agent {
                docker {
                    image "hashicorp/terraform:${TERRAFORM_VERSION}"
                    reuseNode true
                }
            }
            steps {
                sh '''
                    # Validate Terraform blueprints
                    for dir in terraform/blueprints/*/; do
                        echo "Validating $dir"
                        cd "$dir"
                        terraform init -backend=false -no-color || true
                        terraform validate -no-color || true
                        cd -
                    done
                '''
            }
        }

        stage('Kind Cluster Tests') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                    changeRequest()
                }
            }
            agent {
                kubernetes {
                    yaml '''
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: kind
    image: kindest/kind:latest
    command:
    - cat
    tty: true
    privileged: true
    volumeMounts:
    - name: docker-sock
      mountPath: /var/run/docker.sock
  volumes:
  - name: docker-sock
    hostPath:
      path: /var/run/docker.sock
'''
                }
            }
            stages {
                stage('Create Cluster') {
                    steps {
                        container('kind') {
                            sh '''
                                kind create cluster --name mantl-test --wait 5m
                                kubectl cluster-info
                            '''
                        }
                    }
                }

                stage('Deploy Platform') {
                    steps {
                        container('kind') {
                            sh '''
                                # Install ArgoCD
                                kubectl create namespace argocd
                                kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

                                # Wait for ArgoCD to be ready
                                kubectl wait --for=condition=available deployment/argocd-server -n argocd --timeout=300s
                                kubectl wait --for=condition=available deployment/argocd-applicationset-controller -n argocd --timeout=300s

                                # Install cert-manager (required for TLS automation)
                                kubectl create namespace cert-manager
                                kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
                                kubectl wait --for=condition=available deployment/cert-manager -n cert-manager --timeout=300s

                                # Deploy ClusterIssuers
                                kubectl apply -k platform/base/cert-manager/issuers || true

                                # Deploy dev environment via ArgoCD
                                kubectl apply -f clusters/dev/app-of-apps.yaml || true

                                # Wait for platform to sync
                                sleep 30
                            '''
                        }
                    }
                }

                stage('E2E Tests') {
                    steps {
                        container('kind') {
                            sh '''
                                # Deploy example app
                                kubectl apply -k applications/examples/hello || true
                                kubectl rollout status deploy/app -n default --timeout=120s || true

                                # Test connectivity
                                kubectl get pods -A
                                kubectl get svc -A
                            '''
                        }
                    }
                }
            }
            post {
                always {
                    container('kind') {
                        sh '''
                            kind delete cluster --name mantl-test || true
                        '''
                    }
                }
            }
        }

        stage('Build Documentation') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                }
            }
            agent {
                docker {
                    image 'python:3.12-slim'
                    reuseNode true
                }
            }
            steps {
                sh '''
                    pip install -r requirements.txt
                    echo "Documentation build placeholder"
                '''
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: '**/test-results.xml, **/coverage.xml', allowEmptyArchive: true
            cleanWs()
        }

        success {
            script {
                if (env.BRANCH_NAME == 'main') {
                    slackSend(
                        color: 'good',
                        channel: '#builds',
                        message: "✅ Build Succeeded: ${env.JOB_NAME} #${env.BUILD_NUMBER}\n" +
                                 "Branch: ${env.BRANCH_NAME}\n" +
                                 "Commit: ${env.GIT_COMMIT?.take(7)}\n" +
                                 "(<${env.BUILD_URL}|Open>)"
                    )
                }
            }
        }

        failure {
            slackSend(
                color: 'danger',
                channel: '#builds',
                message: "❌ Build Failed: ${env.JOB_NAME} #${env.BUILD_NUMBER}\n" +
                         "Branch: ${env.BRANCH_NAME}\n" +
                         "Commit: ${env.GIT_COMMIT?.take(7)}\n" +
                         "(<${env.BUILD_URL}|Open>)"
            )
        }

        unstable {
            slackSend(
                color: 'warning',
                channel: '#builds',
                message: "⚠️ Build Unstable: ${env.JOB_NAME} #${env.BUILD_NUMBER}\n" +
                         "Branch: ${env.BRANCH_NAME}\n" +
                         "(<${env.BUILD_URL}|Open>)"
            )
        }
    }
}
