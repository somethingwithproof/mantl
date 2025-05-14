pipeline {
    agent {
        docker {
            image 'python:3.10-slim'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }
    
    environment {
        PYTHONPATH = "${WORKSPACE}"
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