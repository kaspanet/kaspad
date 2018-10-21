#!/bin/sh

export ENVIRONMENT_NAME=${ENVIRONMENT_NAME:-"dev"}
export CF_STACK_NAME=${CF_STACK_NAME:-"${ENVIRONMENT_NAME}-ECS-BTCD"}
export SERVICE_NAME=${SERVICE_NAME:-"btcd"}
export IMAGE_TAG=${IMAGE_TAG:-"latest"}
# GIT_COMMIT is set by Jenkins
export COMMIT=${COMMIT:-$GIT_COMMIT}

AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION:-eu-central-1}
export AWS_DEFAULT_REGION
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output=text)
export AWS_ACCOUNT_ID
ECR_SERVER=${ECR_SERVER:-"$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"}
export ECR_SERVER

CF_PARAM=TaskImage
IMAGE_NAME=${ECR_SERVER}/${SERVICE_NAME}

trap "exit 1" INT
fatal() { echo "ERROR: $*" >&2; exit 1; }
tm() {
  START=$(date +%s)
  echo "--> $*" >&2
  "$@"
  rc=$?
  echo "--> took $(($(date +%s) - START))s" >&2
  return $rc
}

test_git_cli() {
  git --version >/dev/null || fatal 'The "git" CLI tool is not available.'
}

test_aws_cli() {
  aws --version >/dev/null || fatal 'The "aws" CLI tool is not available.'
  aws sts get-caller-identity >/dev/null || fatal 'The "aws" CLI tool is not configured.'
}

test_docker_cli() {
  docker --version >/dev/null || fatal 'The "docker" CLI tool is not available.'
}

test_docker_server() {
  docker version -f 'Docker server version {{.Server.Version}}, build {{.Server.GitCommit}}' >/dev/null \
    || fatal 'The "docker" server is not available'
}

# fix $COMMIT if executed without Jenkins
if [ -z "$COMMIT" ]; then
  test_git_cli
  COMMIT=$(git rev-parse --short=7 HEAD)
  export COMMIT
fi

do_version() {
  test_git_cli
  # place environment variables set by Jenkins into a metadata file
  cat <<-EOF > version.txt
	GIT_BRANCH=$BRANCH_NAME
	GIT_COMMIT=$(git rev-parse --short=12 HEAD)
	GIT_AUTHOR_EMAIL=$(git log -1 --pretty='format:%ae')
	GIT_AUTHOR_NAME=$(git log -1 --pretty='format:%an')
	GIT_AUTHOR_DATE=$(git log -1 --pretty='format:%aI')
	EOF
}

do_login() {
  test_aws_cli
  eval "$(aws ecr get-login --no-include-email)"
}

do_build() {
  do_login
  test_docker_cli
  do_version
  tm docker build -t "${SERVICE_NAME}:${COMMIT}" . \
      -f docker/Dockerfile \
      || fatal 'Failed to build the docker image'
}

do_create_ecr() {
    echo "==> Checking for existance of ECR repository..."
    tm aws ecr describe-repositories --query 'repositories[].repositoryName' \
    | grep -E "\"$SERVICE_NAME\"" >/dev/null \
    || {
      echo "==> ECR for $SERVICE_NAME does not exist. Creating ..."
      tm aws ecr create-repository --repository-name "$SERVICE_NAME" \
          || fatal 'Failed to create ECR repository'
    }
}

do_push() {
  test_aws_cli
  test_docker_cli
  test_docker_server
  do_build
  tm docker tag  "${SERVICE_NAME}:${COMMIT}" "${IMAGE_NAME}:${COMMIT}" || fatal 'Failed to tag docker image'
  tm docker tag  "${SERVICE_NAME}:${COMMIT}" "${IMAGE_NAME}:latest" || fatal 'Failed to tag docker image to :last'
  do_create_ecr
  do_login
  tm docker push "${IMAGE_NAME}:${COMMIT}" || fatal 'Failed to push docker image to ECR'
  tm docker push "${IMAGE_NAME}:latest" || fatal 'Failed to push docker image :latest to ECR'
}

do_deploy() {
  tm aws cloudformation \
    update-stack \
    --stack-name "$CF_STACK_NAME" \
    --capabilities CAPABILITY_NAMED_IAM \
    --use-previous-template \
    --parameters "ParameterKey=EnvironmentName,UsePreviousValue=true \
                  ParameterKey=$CF_PARAM,ParameterValue=${IMAGE_NAME}:$COMMIT" \
    || fatal "Failed to update CloudFormation stack $STACK_NAME."
}

do_usage() {
  echo "Usage: $0 <build|login|push|deploy>"
  echo "  version  - create a version.txt file with some meta data"
  echo "  build    - create docker image named $SERVICE_NAME with tag \$COMMIT"
  echo "  login    - configure docker push credentials to use AWS ECR"
  echo "  push     - tag image as :latest and push both :\$COMMIT and :latest to ECR"
  echo "  push_all - push for all AWS regions"
  echo "  deploy   - update CloudFormation stack '$CF_STACK_NAME' with ECR image '${SERVICE_NAME}:${COMMIT}'"
}

do_push_all() {
  for AWS_DEFAULT_REGION in 'us-east-1' 'us-east-2'; do
    export AWS_DEFAULT_REGION
    ECR_SERVER="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"
    export ECR_SERVER
    IMAGE_NAME=${ECR_SERVER}/${SERVICE_NAME}
    export IMAGE_NAME
    do_push
  done
}

case $1 in
  version)  do_version  ;;
  build)    do_build    ;;
  login)    do_login    ;;
  push)     do_push     ;;
  push_all) do_push_all ;;
  deploy)   do_deploy   ;;
  *)        do_usage    ;;
esac
