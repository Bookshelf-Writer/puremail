#!/bin/bash
# Original source: https://github.com/Bookshelf-Writer/scripts-for-integration/blob/main/_run/push-hook.sh
echo "[HOOK]" "Push"

run_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
values_dir="$run_dir/values"
script_dir="$run_dir/scripts"
root_path=$(cd "$run_dir/.." && pwd)

#############################################################################

OLD_VER=$(bash "$script_dir/sys.sh" -v)
VERSION=$(bash "$script_dir/sys.sh" -i -pa)

echo "Updated patch-ver:" "$OLD_VER >> $VERSION"

go mod tidy
#go mod vendor

#############################################################################
exit 0

