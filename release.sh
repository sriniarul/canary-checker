#!/bin/bash
set -x
cd $GITHUB_WORKSPACE
GITHUB_USER=$(echo $GITHUB_REPOSITORY | cut -d/ -f1)
NAME=$(echo $GITHUB_REPOSITORY | cut -d/ -f2)
TAG=$(echo $GITHUB_REF | sed 's|refs/tags/||')
VERSION="$TAG built $(date)"

make static linux darwin compress


github-release release -u $GITHUB_USER -r ${NAME} --tag $TAG || echo Release already created
github-release upload -R -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME} -f .bin/${NAME}
github-release upload -R -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}_osx -f .bin/${NAME}_osx

curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash

cd config
./kustomize edit set image canary-checker:$TAG
./kustomize build  > release.yaml
github-release upload -R -u $GITHUB_USER -r ${NAME} --tag $TAG -n release.yaml -f release.yaml
