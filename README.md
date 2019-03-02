## ao-debug

To compile:
```
1. dep ensure -v
2. go build -o ao-debug main.go
```

The ansible-operator uses ansible-runner to run the ansible code. You need three things to look at the ansible-like-logs of your ansible-operator pod.

1. Operator deployment filepath. A typical ansible-operator project layout will be according to [this](https://github.com/operator-framework/operator-sdk/blob/master/doc/ansible/project_layout.md). Hence default of the operator deployment is set to `./deploy/operator.yaml`
2. Operator Namespace, it is defaulted to `default` namespace
3. Job ID: You could get this by `kubectl logs -l name=etcd-ansible-operator | grep '[[:digit:]]*' -o | tail -1`. The operator pod logs has job id for each task it executes. 

I would recommend using the following command to get the logs:

```
ao-debug --kubeconfig='/home/alpatel/.kube/config' --job-id $(kubectl logs -l name=etcd-ansible-operator | grep '[[:digit:]]*' -o | tail -1)

```

Example logs:

```
[alpatel@alpatel etcd-ansible-operator]$ ao-debug --kubeconfig='/home/alpatel/.kube/config' --job-id
$(kubectl logs -l name=etcd-ansible-operator | grep '[[:digit:]]*' -o | tail -1)
...
...
...

PLAYBOOK: playbook.yaml ********************************************************
1 plays in /opt/ansible/playbook.yaml

PLAY [localhost] ***************************************************************

TASK [Gathering Facts] *********************************************************
task path: /opt/ansible/playbook.yaml:1
ok: [localhost]
META: ran handlers
...
...
...

```

Happy debugging :)
