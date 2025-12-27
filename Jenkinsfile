pipeline {
    agent {
        docker {
            image 'python:3.10-slim'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }
    
    environment {
        PYTHONPATH = "${WORKSPACE}"
        // Toggle to enable Spinnaker trigger stage
        SPIN_TRIGGER_ENABLED = "false" // set to "true" to enable
        // Gate base URL, e.g., https://spinnaker.example.com/gate
        SPIN_GATE_URL = credentials('SPIN_GATE_URL')
        // Jenkins credential ID for a token to authenticate webhook (string)
        SPIN_WEBHOOK_TOKEN = credentials('SPIN_WEBHOOK_TOKEN')
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 1, unit: 'HOURS')
        disableConcurrentBuilds()
    }
    
    triggers {
        pollSCM('H/15 * * * *')
    }
    
    stages {
        stage('Setup') {
            steps {
                sh '''
                    python -m pip install --upgrade pip
                    pip install -r requirements.txt
                    pip install -r requirements-test.txt
                    pip install safety
                '''
            }
        }
        
        stage('Lint') {
            steps {
                sh '''
                    pip install flake8 pylint ansible-lint
                    flake8 --exclude=.git,__pycache__,docs/,old,build,dist
                    pylint --disable=C0111,R0903,C0301 $(find . -name "*.py" | grep -v "__pycache__" | grep -v ".git" | grep -v "docs")
                    ansible-lint roles/* playbooks/*
                '''
            }
        }
        
        stage('Security Scan') {
            steps {
                sh '''
                    safety scan -r requirements.txt -r requirements-test.txt
                '''
            }
        }
        
        stage('Test') {
            steps {
                sh '''
                    pytest tests/ --junitxml=test-results.xml --cov=. --cov-report=xml
                '''
            }
            post {
                always {
                    junit 'test-results.xml'
                    recordCoverage(tools: [[parser: 'COBERTURA', pattern: 'coverage.xml']])
                }
            }
        }
        
        stage('Build Documentation') {
            steps {
                sh '''
                    cd docs
                    make html
                '''
            }
            post {
                success {
                    publishHTML(target: [
                        allowMissing: false,
                        alwaysLinkToLastBuild: false,
                        keepAll: true,
                        reportDir: 'docs/_build/html',
                        reportFiles: 'index.html',
                        reportName: 'Documentation'
                    ])
                }
            }
        }
        
        stage('Integration Tests') {
            when {
                branch 'main'
            }
            steps {
                sh '''
                    cd tests/integration/kubernetes-nomad/test
                    python standalone-test.py -v
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
                    -d '{"artifacts":[]}'
                '''
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
        success {
            slackSend(
                color: 'good',
                message: "Build Succeeded: ${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
            )
        }
        failure {
            slackSend(
                color: 'danger',
                message: "Build Failed: ${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
            )
        }
    }
}