#!/bin/sh

export ENVIRONMENT_NAME=${ENVIRONMENT_NAME:-"dev"}
export CF_STACK_NAME=${CF_STACK_NAME:-"${ENVIRONMENT_NAME}-ECS-BTCD"}
export SERVICE_NAME=${SERVICE_NAME:-"btcd"}
export IMAGE_TAG=${IMAGE_TAG:-"latest"}
# GIT_COMMIT is set by Jenkins
export COMMIT=${COMMIT:-$GIT_COMMIT}

export AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION:-eu-central-1}
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output=text)
export ECR_SERVER=${ECR_SERVER:-"$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"}

CF_PARAM=TaskImage
IMAGE_NAME=${ECR_SERVER}/${SERVICE_NAME}

notify_telegram() {
  # Wait for the process to finish so it could flush logs etc.
  sleep 10s

  # Build the failure message
  MESSAGE="*${ghprbActualCommitAuthor}*:
Build *FAILED* for pull request '${ghprbPullTitle}'
[Github](${ghprbPullLink})        [Jenkins](${BUILD_URL}console)"

  # Send the failure message
  curl -s \
    -X POST \
    "https://api.telegram.org/bot${TELEGRAM_API_TOKEN}/sendMessage" \
    -d chat_id="${TELEGRAM_CHAT_ID}" \
    -d parse_mode=markdown \
    -d disable_web_page_preview=true \
    -d text="${MESSAGE}"

  # Retrieve the build log
  LOG=$(curl ${BUILD_URL}consoleText)

  # Send the build log
  printf "$LOG" | curl \
    "https://api.telegram.org/bot${TELEGRAM_API_TOKEN}/sendDocument" \
    -F chat_id="${TELEGRAM_CHAT_ID}" \
    -F document="@-;filename=build.log"
}

trap "exit 1" INT
fatal() {
  echo "ERROR: $*" >&2
  notify_telegram &

  exit 1
}

measure_runtime() {
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

version() {
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

login() {
  test_aws_cli
  eval "$(aws ecr get-login --no-include-email)"
}

build() {
  login
  test_docker_cli
  version
  measure_runtime docker build -t "${SERVICE_NAME}:${COMMIT}" . \
      -f docker/Dockerfile \
      || fatal 'Failed to build the docker image'
}

create_ecr() {
    echo "==> Checking for existance of ECR repository..."
    measure_runtime aws ecr describe-repositories --query 'repositories[].repositoryName' \
    | grep -E "\"$SERVICE_NAME\"" >/dev/null \
    || {
      echo "==> ECR for $SERVICE_NAME does not exist. Creating ..."
      measure_runtime aws ecr create-repository --repository-name "$SERVICE_NAME" \
          || fatal 'Failed to create ECR repository'
    }
}

push() {
  test_aws_cli
  test_docker_cli
  test_docker_server
  build
  measure_runtime docker tag  "${SERVICE_NAME}:${COMMIT}" "${IMAGE_NAME}:${COMMIT}" || fatal 'Failed to tag docker image'
  measure_runtime docker tag  "${SERVICE_NAME}:${COMMIT}" "${IMAGE_NAME}:latest" || fatal 'Failed to tag docker image to :last'
  create_ecr
  login
  measure_runtime docker push "${IMAGE_NAME}:${COMMIT}" || fatal 'Failed to push docker image to ECR'
  measure_runtime docker push "${IMAGE_NAME}:latest" || fatal 'Failed to push docker image :latest to ECR'
}

deploy() {
  measure_runtime aws cloudformation \
    update-stack \
    --stack-name "$CF_STACK_NAME" \
    --capabilities CAPABILITY_NAMED_IAM \
    --use-previous-template \
    --parameters "ParameterKey=EnvironmentName,UsePreviousValue=true \
                  ParameterKey=$CF_PARAM,ParameterValue=${IMAGE_NAME}:$COMMIT" \
    || fatal "Failed to update CloudFormation stack $STACK_NAME."
}

usage() {
  echo "Usage: $0 <build|login|push|deploy>"
  echo "  version  - create a version.txt file with some meta data"
  echo "  build    - create docker image named $SERVICE_NAME with tag \$COMMIT"
  echo "  login    - configure docker push credentials to use AWS ECR"
  echo "  push     - tag image as :latest and push both :\$COMMIT and :latest to ECR"
  echo "  push_all - push for all AWS regions"
  echo "  deploy   - update CloudFormation stack '$CF_STACK_NAME' with ECR image '${SERVICE_NAME}:${COMMIT}'"
}

push_all() {
  for AWS_DEFAULT_REGION in 'us-east-1' 'us-east-2'; do
    export AWS_DEFAULT_REGION
    ECR_SERVER="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"
    export ECR_SERVER
    IMAGE_NAME=${ECR_SERVER}/${SERVICE_NAME}
    export IMAGE_NAME
    push
  done
}

case $1 in
  version)  version  ;;
  build)    build    ;;
  login)    login    ;;
  push)     push     ;;
  push_all) push_all ;;
  deploy)   deploy   ;;
  *)        usage    ;;
esac
