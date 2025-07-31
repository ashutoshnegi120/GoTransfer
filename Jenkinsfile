pipeline {
    agent any

    stages {
        stage('Checkout') {
            steps {
                // Pull the code from the configured SCM (e.g., Git)
                checkout scm
            }
        }
        stage('Build') {
            steps {
                // Run go build to build the project 
                sh 'go build -o go-transfer ./cmd/Go_transfer'
            }
        }
        stage('Test') {
            steps {
                // Run your test suite (example: Maven tests)
                sh 'go test ./cmd/Go_transfer/ -args -config=../../.config/config.yml'
            }
        }
    }
    post {
        success {
            emailext subject: "✅ Build SUCCESS: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                     body: """Good news!

Build *#${env.BUILD_NUMBER}* of *${env.JOB_NAME}* completed successfully.

Check it here: ${env.BUILD_URL}""",
                     to: "ashutoshnegisgrr@gmail.com"
        }

        failure {
            emailext subject: "❌ Build FAILED: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                     body: """Oops.

Build *#${env.BUILD_NUMBER}* of *${env.JOB_NAME}* failed.

Check console output: ${env.BUILD_URL}console""",
                     to: "ashutoshnegisgrr@gmail.com"
        }

        always {
            echo 'Build completed...'
        }
    }
}
