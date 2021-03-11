#!/usr/bin/env bash
#_CMDS az
#_PARAMS clusterName resourceGroup

az aks get-credentials -n $clusterName -g $resourceGroup --overwrite-existing
