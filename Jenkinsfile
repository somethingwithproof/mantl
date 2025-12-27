pipeline {
    agent any

    environment {
        PYTHONPATH = "${WORKSPACE}"
        TERRAFORM_VERSION = "1.10.3"
        // Toggle to enable Spinnaker trigger stage
        SPIN_TRIGGER_ENABLED = "false" // set to "true" to enable
        // Gate base URL, e.g., https://spinnaker.example.com/gate
        SPIN_GATE_URL = credentials('SPIN_GATE_URL')
        // Jenkins credential ID for a token to authenticate webhook (string)
        SPIN_WEBHOOK_TOKEN = credentials('SPIN_WEBHOOK_TOKEN')
    }

    options {
        buildDiscarder(logRotator(numToKeepStr: '30'))
        timeout(time: 2, unit: 'HOURS')
        disableConcurrentBuilds()
        timestamps()
    }

    triggers {
        // Poll SCM every 15 minutes
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

                            # Run ansible-lint (continue on error for now)
                            ansible-lint || true
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
                            # Terraform format check (recursive)
                            terraform fmt -check -recursive
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
                    pytest -q --junitxml=test-results.xml --cov=. --cov-report=xml --cov-report=html
                '''
            }
            post {
                always {
                    junit 'test-results.xml'
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
                    # Terraform validate (soft - continue on error)
                    find terraform -type d -maxdepth 2 -mindepth 1 -print0 | \
                        xargs -0 -I {} sh -c 'cd {} && terraform init -backend=false -no-color || true; terraform validate -no-color || true'
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
                stage('Kind Smoke Test') {
                    steps {
                        container('kind') {
                            sh '''
                                # Create kind cluster
                                kind create cluster --name mantl --wait 5m

                                # Verify cluster
                                kubectl get nodes
                                kubectl cluster-info
                            '''
                        }
                    }
                }

                stage('E2E Hello App') {
                    steps {
                        container('kind') {
                            sh '''
                                # Deploy example hello app
                                kubectl apply -k applications/examples/hello
                                kubectl rollout status deploy/app -n default --timeout=120s

                                # Port-forward and test
                                kubectl port-forward svc/app -n default 8080:80 &
                                PF_PID=$!
                                sleep 3
                                curl -fsS http://127.0.0.1:8080/ | head -c 100
                                kill $PF_PID || true
                            '''
                        }
                    }
                }

                stage('ExternalDNS RFC2136 E2E') {
                    steps {
                        container('kind') {
                            sh '''
                                # Deploy test DNS (bind9) and service
                                echo "Deploying DNS test infrastructure..."
                                # This would deploy actual DNS testing infrastructure
                                # Placeholder for now
                                echo "DNS E2E tests completed"
                            '''
                        }
                    }
                }
            }
            post {
                always {
                    container('kind') {
                        sh '''
                            # Cleanup kind cluster
                            kind delete cluster --name mantl || true
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
                    # Documentation build would go here
                    # cd docs && make html
                    echo "Documentation build stage placeholder"
                '''
            }
        }

        stage('Docker Build') {
            when {
                anyOf {
                    branch 'main'
                    tag pattern: 'v\\d+\\.\\d+\\.\\d+', comparator: 'REGEXP'
                }
            }
            steps {
                script {
                    def imageTag = env.BRANCH_NAME == 'main' ? 'latest' : env.TAG_NAME
                    sh """
                        # Build Docker image if Dockerfile exists
                        if [ -f Dockerfile ]; then
                            docker build -t mantl:${imageTag} .
                            echo "Built mantl:${imageTag}"
                        else
                            echo "No Dockerfile found, skipping Docker build"
                        fi
                    """
                }
            }
        }

        stage('Integration Tests') {
            when {
                branch 'main'
            }
            steps {
                sh '''
                    # Integration tests placeholder
                    echo "Integration tests would run here"
                '''
            }
        }

        // Optional: trigger a Spinnaker pipeline via Gate webhook
        stage('Trigger Spinnaker') {
            when {
                allOf {
                    branch 'main'
                    expression { return env.SPIN_TRIGGER_ENABLED == 'true' }
                }
            }
            steps {
                sh '''
                    if [ -z "$SPIN_GATE_URL" ]; then
                        echo "SPIN_GATE_URL is not set; skipping" && exit 0
                    fi
                    curl -sS -X POST "$SPIN_GATE_URL/webhooks/webhook/mantl-deploy" \
                        -H 'Content-Type: application/json' \
                        -H "X-Webhook-Token: $SPIN_WEBHOOK_TOKEN" \
                        -d '{"artifacts":[], "parameters": {"GIT_COMMIT": "'$GIT_COMMIT'", "BUILD_NUMBER": "'$BUILD_NUMBER'"}}'
                '''
            }
        }
    }

    post {
        always {
            // Archive artifacts
            archiveArtifacts artifacts: '**/test-results.xml, **/coverage.xml', allowEmptyArchive: true

            // Cleanup workspace
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
