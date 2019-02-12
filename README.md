## ao-debug

To compile:
```
1. dep ensure -v
2. go build -o ao-debug main.go
```

The ansible-operator uses ansible-runner to run the ansible code. You need three things to look at the ansible-like-logs of your ansible-operator pod.

1. Operator Pod name
2. Operator Namespace
3. Job ID: You could get this by `kubectl logs <operator-pod>`. The operator pod logs has job id for each task it executes. 

Run the ao-debug with following command `./ao-debug --pod-name <pod-name> --pod-namespace <namespace> --job-id <job-id>`. By default the job id would be set to `latest`.

Example logs:

```
[alpatel@alpatel ao-debug]$ ./ao-debug --pod-name=ansible-service-broker-operator-779bdd4947-5h
56j --pod-namespace=fail --job-id=2376680986980114193
Request URL: https://alpatel-dev-api.devcluster.openshift.com:6443/api/v1/namespaces/fail/pods/ansible-service-broker-operator-779bdd4947-5h56j/exec?command=tail&command=-f&command=-n&command=%2B0&command=%2Ftmp%2Fansible-operator%2Frunner%2Fosb.openshift.io%2Fv1alpha1%2FAutomationBroker%2Ffail%2Fansible-service-broker%2Fartifacts%2F2376680986980114193%2Fstdout&container=automation-broker-operator&stderr=true&stdin=true&stdout=true
ansible-playbook 2.7.7
  config file = /etc/ansible/ansible.cfg
  configured module search path = [u'/usr/share/ansible/openshift']
  ansible python module location = /usr/lib/python2.7/site-packages/ansible
  executable location = /usr/bin/ansible-playbook
  python version = 2.7.5 (default, Oct 30 2018, 23:45:53) [GCC 4.8.5 20150623 (Red Hat 4.8.5-36)]
Using /etc/ansible/ansible.cfg as config file

/tmp/ansible-operator/runner/osb.openshift.io/v1alpha1/AutomationBroker/fail/ansible-service-broker/inventory/hosts did not meet host_list requirements, check plugin documentation if this is unexpected
/tmp/ansible-operator/runner/osb.openshift.io/v1alpha1/AutomationBroker/fail/ansible-service-broker/inventory/hosts did not meet script requirements, check plugin documentation if this is unexpected

statically imported: /opt/ansible/roles/ansible-service-broker/tasks/validate_present.yml

statically imported: /opt/ansible/roles/ansible-service-broker/tasks/tls_k8s.yml

statically imported: /opt/ansible/roles/ansible-service-broker/tasks/tls_k8s.yml

PLAYBOOK: playbook.yaml ********************************************************
1 plays in /opt/ansible/playbook.yaml

PLAY [localhost] ***************************************************************
META: ran handlers

...

```

`Note: Press Ctrl+C to exit` 

Happy debugging :)
