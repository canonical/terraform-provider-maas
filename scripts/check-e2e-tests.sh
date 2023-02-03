#!/bin/bash

RESULTS_BRANCH="main"
RESULTS_URL="https://raw.githubusercontent.com/maas/maas-terraform-e2e-tests/${RESULTS_BRANCH}/results.json"

function main () {
    local results="$(curl -L $RESULTS_URL)"
    local failures="$(echo $results | jq '.results[0].testsuites.testsuite["@failures"]')"
    local errors="$(echo $results | jq '.results[0].testsuites.testsuite["@errors"]')"
    ([ "${failures}" == "0" ] && [ "${errors}" == "0" ]) || (echo "${failures} failures and ${errors} errors found" && exit 1)
}

main
