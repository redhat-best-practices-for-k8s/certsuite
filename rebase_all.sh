#!/bin/bash

# Update the remotes for good measure
git remote update
git fetch

# Update the main branches first
MAIN_BRANCH=main
git checkout $MAIN_BRANCH
git pull --rebase upstream $MAIN_BRANCH
git push origin $MAIN_BRANCH

# Loop through all of the local branches and rebase them.
for i in $(git branch | grep -v $MAIN_BRANCH); do git checkout $i; git pull --rebase upstream $MAIN_BRANCH; git checkout $MAIN_BRANCH; git branch -d $i; done
