#!/bin/bash

if [ "$1" = "-coverage" ]; then
  echo 'Will output test coverage at end'
  CHECK_COVERAGE=true
  COVERMODE='set'
  echo "mode: $COVERMODE" >profile.cov
fi

for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d -not -path './vendor/*'); do
  if ls "$dir"/*.go &>/dev/null; then
    if [ "$CHECK_COVERAGE" = true ]; then
      go test -short -covermode="$COVERMODE" -coverprofile="$dir"/profile.tmp "$dir"
      if [ -f "$dir"/profile.tmp ]; then
        cat "$dir"/profile.tmp | tail -n +2 >>profile.cov
        rm "$dir"/profile.tmp
      fi
    else
      go test -short "$dir"
    fi
  fi
done

if [ "$CHECK_COVERAGE" = true ]; then
  go tool cover -func profile.cov
  go tool cover -html=profile.cov
fi
