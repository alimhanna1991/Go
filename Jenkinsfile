pipeline {
    agent any
    
    environment {
        // Application configuration
        APP_NAME = 'webpage-analyzer'
        GO_VERSION = '1.21'
        APP_DIR = 'webpage-analyzer'
        
        // Docker configuration
        DOCKER_REGISTRY = 'docker.io'
        DOCKER_IMAGE = "${DOCKER_REGISTRY}/${env.JOB_NAME}"
        DOCKER_TAG = "${env.BUILD_NUMBER}"
        
        // Test configuration
        COVERAGE_THRESHOLD = '80'
        
        // Go environment
        GOPATH = "${env.WORKSPACE}/.go"
        GOCACHE = "${env.WORKSPACE}/.cache/go-build"
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        ansiColor('xterm')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    echo "Checking out ${env.GIT_URL} - ${env.GIT_BRANCH}"
                    env.GIT_COMMIT_SHORT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
            }
        }
        
        stage('Setup') {
            steps {
                dir("${APP_DIR}") {
                    sh '''
                        mkdir -p ../.go ../.cache/go-build
                        export GOPATH=${WORKSPACE}/.go
                        export GOCACHE=${WORKSPACE}/.cache/go-build
                        go version
                        go mod download
                        go mod verify
                    '''
                }
            }
        }
        
        stage('Lint') {
            steps {
                dir("${APP_DIR}") {
                    sh '''
                        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
                            sh -s -- -b $(go env GOPATH)/bin v1.54.2
                        golangci-lint run ./... --out-format=colored-line-number
                    '''
                }
            }
            post {
                always {
                    recordIssues(tools: [golangciLint()])
                }
            }
        }
        
        stage('Unit Tests') {
            steps {
                dir("${APP_DIR}") {
                    sh '''
                        go test -v -coverprofile=coverage.out ./...
                        go tool cover -html=coverage.out -o coverage.html
                        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
                        echo "Coverage: $COVERAGE%"
                        if (( $(echo "$COVERAGE < ${COVERAGE_THRESHOLD}" | bc -l) )); then
                            echo "Coverage $COVERAGE% is below threshold ${COVERAGE_THRESHOLD}%"
                            exit 1
                        fi
                    '''
                }
            }
            post {
                always {
                    publishHTML([
                        reportDir: "${APP_DIR}",
                        reportFiles: 'coverage.html',
                        reportName: 'Go Coverage Report',
                        reportTitles: 'Test Coverage'
                    ])
                }
            }
        }
        
        stage('Integration Tests') {
            when {
                branch 'main'
                expression { env.CHANGE_ID == null }
            }
            steps {
                script {
                    docker.withRegistry("https://${DOCKER_REGISTRY}") {
                        def app = docker.build("${DOCKER_IMAGE}:${DOCKER_TAG}", "-f Dockerfile .")
                        app.withRun('-p 8080:8080') { container ->
                            sh '''
                                echo "Waiting for service to be ready..."
                                for i in {1..30}; do
                                    if curl -s http://localhost:8080 > /dev/null; then
                                        break
                                    fi
                                    sleep 1
                                done
                                curl -fsS http://localhost:8080/ || exit 1
                                curl -fsS -X POST -d "url=https://example.com" http://localhost:8080/analyze || exit 1
                            '''
                        }
                    }
                }
            }
        }
        
        stage('Security Scan') {
            when {
                branch 'main'
            }
            steps {
                dir("${APP_DIR}") {
                    sh '''
                        go install github.com/securego/gosec/v2/cmd/gosec@latest
                        gosec -fmt=html -out=security-report.html ./...
                    '''
                }
            }
            post {
                always {
                    publishHTML([
                        reportDir: "${APP_DIR}",
                        reportFiles: 'security-report.html',
                        reportName: 'Security Scan Report'
                    ])
                }
            }
        }
        
        stage('Build Binary') {
            steps {
                dir("${APP_DIR}") {
                    sh '''
                        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
                            -o ${APP_NAME} \
                            ./main.go
                        ls -lh ${APP_NAME}
                    '''
                }
            }
            post {
                success {
                    archiveArtifacts artifacts: "${APP_DIR}/${APP_NAME}", fingerprint: true
                }
            }
        }
        
        stage('Build Docker Image') {
            steps {
                script {
                    docker.withRegistry("https://${DOCKER_REGISTRY}") {
                        // Build production image
                        def prodImage = docker.build("${DOCKER_IMAGE}:${DOCKER_TAG}", "-f Dockerfile .")
                        prodImage.push()
                        prodImage.push('latest')
                        
                        // Build development image for develop branch
                        if (env.BRANCH_NAME == 'develop') {
                            def devImage = docker.build("${DOCKER_IMAGE}:dev-${DOCKER_TAG}", "-f Dockerfile.dev .")
                            devImage.push()
                        }
                    }
                }
            }
        }
        
        stage('Deploy to Staging') {
            when {
                branch 'main'
                expression { env.CHANGE_ID == null }
            }
            steps {
                script {
                    sh '''
                        # Deploy to staging environment
                        ssh staging-server << EOF
                            cd /opt/${APP_NAME}
                            docker pull ${DOCKER_IMAGE}:${DOCKER_TAG}
                            docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:staging
                            docker compose -f docker-compose.prod.yml up -d --remove-orphans
                            docker system prune -f
                        EOF
                    '''
                }
            }
        }
        
        stage('Smoke Tests') {
            when {
                branch 'main'
            }
            steps {
                script {
                    sh '''
                        # Wait for deployment
                        sleep 10
                        
                        # Run smoke tests
                        curl -f http://staging-server:8080 || exit 1
                        curl -f http://staging-server:8080/ || exit 1
                        
                        # Test URL analysis
                        curl -X POST -d "url=https://example.com" http://staging-server:8080/analyze || exit 1
                    '''
                }
            }
        }
        
        stage('Deploy to Production') {
            when {
                branch 'main'
                input message: 'Deploy to production?', ok: 'Yes'
            }
            steps {
                script {
                    sh '''
                        # Deploy to production environment
                        ssh prod-server << EOF
                            cd /opt/${APP_NAME}
                            docker pull ${DOCKER_IMAGE}:${DOCKER_TAG}
                            docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:production
                            docker compose -f docker-compose.prod.yml up -d --remove-orphans
                            docker system prune -f
                        EOF
                    '''
                }
            }
        }
    }
    
    post {
        always {
            cleanWs()
            script {
                // Clean up old Docker images
                sh '''
                    docker image prune -f
                    docker system prune -f
                '''
            }
        }
        success {
            script {
                // Send Slack notification
                slackSend(
                    color: 'good',
                    message: """
                        ✅ Build ${env.BUILD_NUMBER} succeeded!
                        • Project: ${env.JOB_NAME}
                        • Branch: ${env.GIT_BRANCH}
                        • Commit: ${env.GIT_COMMIT_SHORT}
                        • Duration: ${currentBuild.durationString}
                        • URL: ${env.BUILD_URL}
                    """.stripIndent()
                )
                
                // Send email notification
                emailext(
                    subject: "SUCCESS: ${env.JOB_NAME} - Build ${env.BUILD_NUMBER}",
                    body: "Build succeeded! Check the console output at ${env.BUILD_URL}",
                    to: 'team@example.com'
                )
            }
        }
        failure {
            script {
                slackSend(
                    color: 'danger',
                    message: """
                        ❌ Build ${env.BUILD_NUMBER} failed!
                        • Project: ${env.JOB_NAME}
                        • Branch: ${env.GIT_BRANCH}
                        • Commit: ${env.GIT_COMMIT_SHORT}
                        • URL: ${env.BUILD_URL}
                    """.stripIndent()
                )
                
                emailext(
                    subject: "FAILURE: ${env.JOB_NAME} - Build ${env.BUILD_NUMBER}",
                    body: "Build failed! Check the console output at ${env.BUILD_URL}",
                    to: 'team@example.com'
                )
            }
        }
        unstable {
            script {
                slackSend(
                    color: 'warning',
                    message: "⚠️ Build ${env.BUILD_NUMBER} is unstable: ${env.JOB_NAME}"
                )
            }
        }
    }
}
