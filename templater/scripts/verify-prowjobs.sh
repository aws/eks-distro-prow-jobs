set -e
set -u
set -o pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)

DIFF_LINE_COUNT=$(git diff $REPO_ROOT/jobs | wc -l)
if [ $DIFF_LINE_COUNT -ne 0 ]; then
    CHANGED_FILES=$(git diff --name-only $REPO_ROOT/jobs)
    git diff $REPO_ROOT/jobs
    echo "\n❌ Detected discrepancies between generated and expected Prowjobs!"
    echo "The following generated files need to be checked in:\n"
    echo "${CHANGED_FILES}\n" | tr ' ' '\n'
    exit 1
fi
