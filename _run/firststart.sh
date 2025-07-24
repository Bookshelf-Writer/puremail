#!/bin/bash

scripts/git.sh --add_commit
scripts/git.sh --add_push

cd ../

go mod tidy