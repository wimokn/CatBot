#!/bin/bash

APP_VERSION=1.0.8
DEPLOY_ENV=prod
DOCKER_REPOSITORY=app/bot
DOCKER_SERVER=registry.dreamyard.dev

# 2. commit will push docker image to repository
function commit() {
    local IMAGE=$1
    echo "docker push image : $DOCKER_REPOSITORY/$IMAGE:$DEPLOY_ENV-$APP_VERSION"
    docker push $DOCKER_SERVER/$DOCKER_REPOSITORY/$IMAGE:$DEPLOY_ENV-$APP_VERSION
}

# 3. build_api is the main function to build Dockerfile
function build_api() {
    local IMAGE=catbot
    
    # If found go in default path, it will use go from default path
    GO=/usr/local/go/bin/go
    if [ -f "$GO" ]; then
        /usr/local/go/bin/go mod init oms-proxy
        /usr/local/go/bin/go get
        /usr/local/go/bin/go mod vendor
    else 
        go mod init oms-proxy
        go get
        go mod vendor
    fi

    docker build --platform linux/arm64/v8 -f Dockerfile.arm64 -t $DOCKER_SERVER/$DOCKER_REPOSITORY/$IMAGE:$DEPLOY_ENV-$APP_VERSION .
    commit $IMAGE
}

# 4. Validate APP_VERSION must not empty
if [ "$APP_VERSION" = "" ]; then
    echo -e "APP_VERSION cannot be blank"
    exit 1
fi

# 6. Validate DEPLOY_ENV must not empty
if [ "$DEPLOY_ENV" == "" ]; then
    echo -e "DEPLOY_ENV cannot be blank"
    exit 1
fi

# 7. Validate DOCKER_REPOSITORY must not empty
if [ "$DOCKER_REPOSITORY" == "" ]; then
    echo -e "DOCKER_REPOSITORY cannot be blank"
    exit 1
fi

# 8. Run main build process
build_api