#!/bin/sh

#
# run_vault:
#   A function that executes a docker run command to run a container with the
#   Vault image in the background.
#
# params:
#   1: port number - Specifies the port number to use in the container and on
#                    the host. Defaults to 8200.
#
function run_vault {
  docker run \
      -d \
      -p ${1:-8200}:${1:-8200} \
      vault:latest \
      vault \
      server \
      -dev \
      -dev-root-token-id=root \
      -dev-listen-address=0.0.0.0:${1:-8200}
}

#
# check_vault_ready:
#   A function that executes the docker run command to run the vault status
#   command in a Docker container inside of a loop for up to 120 seconds.  If
#   the command succeeds the function immediately returns a success code (0).
#   Otherwise, the loop re-iterates until 120 seconds have elapsed, at which
#   point the function returns an error code (1).
#
# params:
#   1: port number - The port number to use in the Vault address.
#
function check_vault_ready {
  # Reset timer
  SECONDS=0

  while [ ${SECONDS} -lt 120 ]; do
    if curl -fs http://localhost:${1:-8200}/v1/sys/seal-status > /dev/null ; then
      return 0
    fi    
  done

  return 1
}

# Run the target Vault server
target_container=$(run_vault 8200)

# Run the source Vault server
source_container=$(run_vault 8300)

# Make sure the target Vault server is ready
if ! check_vault_ready 8200 ; then
  echo "Target Vault server is NOT ready"
  exit 1
fi

# Make sure the source Vault server is ready
if ! check_vault_ready 8300 ; then
  echo "Source Vault server is NOT ready"
  exit 1
fi

terraform init
terraform apply -auto-approve

export TARGET_VAULT_TOKEN=$(curl -fs -H "X-Vault-Token: root" -d '{"policies":["hvc-kv"],"ttl":"15m"}' http://localhost:8200/v1/auth/token/create | jq -r '.auth.client_token')
export SOURCE_VAULT_TOKEN_1=$(curl -fs -H "X-Vault-Token: root" -d '{"policies":["hvc-kv1"],"ttl":"15m"}' http://localhost:8300/v1/auth/token/create | jq -r '.auth.client_token')
export SOURCE_VAULT_TOKEN_2=$(curl -fs -H "X-Vault-Token: root" -d '{"policies":["hvc-kv2"],"ttl":"15m"}' http://localhost:8300/v1/auth/token/create | jq -r '.auth.client_token')

failed=0

# Run specifications expected to succeed
echo ""
for f in successful/*.json ; do
  if ! go run ../cmd/hvc/main.go copy $f ; then
    echo "FAIL - Testcase $f"
    failed=$((failed+1))
  else
    echo "PASS - Testcase $f"
  fi
done
echo ""

# Run specifications expected to encounter an error
# for f in unsuccessful/*.json ; do
#   if go run ../cmd/hvc/main.go copy $f ; then
#     echo "FAIL - Testcase $f"
#     failed=$((failed+1))
#   else
#     echo "PASS - Testcase $f"
#   fi
# done

#read -p "Press ENTER to unblock and destroy..." answer

terraform destroy -auto-approve

docker rm -f ${target_container} ${source_container}

if [ $failed -gt 0 ] ; then
  exit 1
fi
