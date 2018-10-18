node {
    stage 'Checkout'
    checkout scm

    stage 'Version'
    sh './deploy.sh version'

    stage 'Build'
    sh "./deploy.sh build"
}
