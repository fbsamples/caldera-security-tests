#!/bin/bash
set -ex

pkg=$(go list ./...)
for dir in */; do
    if [[ "${dir}" != ".mage" ]] \
                              && [[ "${dir}" != "config/" ]] \
                              && [[ "${dir}" != "cmd/" ]] \
                              && [[ "${dir}" != "bin/" ]] \
                              && [[ "${dir}" != "images/" ]] \
                           && [[ "${dir}" != "resources/" ]] \
                                 && [[ "${dir}" != "docs/" ]] \
                            && [[ "${dir}" != "files/" ]] \
                             && [[ "${dir}" != "logs/" ]]; then
        go vet "${pkg}/${dir}"
    fi
done
