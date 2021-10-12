#!/bin/bash
echo ":::setVar cheese=\"ere $kubeConfig assas\""

$kubeConfig=${kubeConfig:-$target_name}

echo az aks get-credentials -n $clusterName -g $resourceGroup --overwrite-existing --file $kubeConfig

echo ":::setParam fo_o=\"wibble $kubeConfig\""