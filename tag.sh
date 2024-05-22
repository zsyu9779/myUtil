#!/bin/bash
git checkout main
git pull origin main
git push origin main
tagVersion=0.0.1
fullCommit=`git rev-parse HEAD`
shortCommit=${fullCommit: 0: 12}
date=`git log --pretty=format:"%ad" $fullCommit -1 --date=format:'%Y%m%d%H%M%S'`
tag=v$tagVersion-$date-$shortCommit
echo "new tag $tag"
echo "tagging"
git tag $tag
echo "push to origin"
git push origin $tag
