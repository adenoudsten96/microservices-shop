terraform apply -auto-approve
ansible-playbook --vault-password-file vault_pass.txt deploy_cluster.yml
kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/calico.yaml
kubectl apply -f ../kubernetes/

echo "Infrastructure deployed."
