#!/bin/bash

if [[ "$mode" == "identity" ]]; then
  az login --identity
elif [[ "$mode" == "servicePrincipal" ]]; then
  echo "az login --tenant "$tenant" --service-principal --username "$clientId" --password "$clientSecret""
else
  echo "Unknown mode supplied, should be 'identity' or 'servicePrincipal'"
  exit 1
fi