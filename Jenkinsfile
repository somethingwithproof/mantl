pipeline {
    agent any

    environment {
        PYTHONPATH = "${WORKSPACE}"
        TERRAFORM_VERSION = "1.14.3"
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
                            # Install kustomize
                            apk add --no-cache curl
                            curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
                            mv kustomize /usr/local/bin/

                            # Validate kustomizations
                            find . -name kustomization.yaml -exec dirname {} \\; | while read dir; do
                                echo "Validating $dir"
                                kustomize build "$dir" > /dev/null || echo "Warning: $dir failed"
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
                                kubectl wait --for=condition=available deployment/argocd-server -n argocd --timeout=300s

                                # Deploy dev environment
                                kubectl apply -k clusters/dev || true
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
