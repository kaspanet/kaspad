node {
    stage 'Checkout'
    checkout scm

    stage 'Version'
    sh './deploy.sh version'

    stage 'Build'
    sh "./deploy.sh build"

    stage 'Push Docker'
    sh "./deploy.sh push"

    stage 'Integration Test'
    echo 'Starting integration test....'
}
