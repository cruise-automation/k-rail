# Script to scan container vulnerabilitys
export VAULT_ADDR=https://vault.secure.car:8200
TL_USER=$(vault read -field=user secret/application/global/default/twistcli-scanner) && export TL_USER
TL_PASS=$(vault read -field=pass secret/application/global/default/twistcli-scanner) && export TL_PASS

# # When updating to Version 19.03 this endpoint is avaliable
# TOKEN=$(echo -n "$TL_USER:$TL_PASS" | openssl base64)
# VERSION=$( curl -sSL -k -H 'Authorization: Basic $TOKEN' \
#  $CONSOLE/api/v1/version | jq -r '.version' )
VERSION="18_11_128"
SCANNER="gcr.io/cruise-gcr-dev/tl_scan:$VERSION"
# FAIL_THRESHOLD="high" # Add when ready to block repo push

# Start scanning for issues
docker run -v /var/run/docker.sock:/var/run/docker.sock \
  --env TL_USER --env TL_PASS \
  $SCANNER $1
