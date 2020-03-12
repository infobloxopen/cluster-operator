# Cluster Operator
Project that provisions kuberneres (k8s) cluster using k8s 
[operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

The project is build around 
[Operator SDK Framework](https://github.com/operator-framework/operator-sdk).
Reference its
[Prerequisites](https://github.com/operator-framework/operator-sdk#prerequisites) 
and how to
[install Operator SDK CLI](https://github.com/operator-framework/operator-sdk#install-the-operator-sdk-cli)
if you are going to extend this project.

## Prerequisite

   * [kops](https://github.com/kubernetes/kops)

## Development

The base project was created using:
```bash
operator-sdk new cluster-operator
cd cluster-operator
operator-sdk add api --api-version=cluster-operator.infobloxopen.github.com/v1alpha1 --kind=Cluster
operator-sdk add controller  --api-version=cluster-operator.infobloxopen.github.com/v1alpha1 --kind=Cluster
```

You can use 
[kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
or 
[minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) for development
```bash
kind create cluster
```

The Operator SDK does a lot of the heavy lifting we can focus on custom type definition for our
Cluster object in 
[cluster_types.go](https://github.com/infobloxopen/cluster-operator/blob/master/pkg/apis/clusteroperator/v1alpha1/cluster_types.go)
and the business logic in
[cluster_controller.go](https://github.com/infobloxopen/cluster-operator/blob/master/pkg/controller/cluster/cluster_controller.go).

Ater making changes run:
```bash
operator-sdk generate k8s
```
regenerate code using code-gen code captured in
[zz_generated.deepcopy.go](https://github.com/infobloxopen/cluster-operator/blob/master/pkg/apis/clusteroperator/v1alpha1/zz_generated.deepcopy.go)

You can do custom validation using
[kubebuilder tags](https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html)

### Local Testing

#### Initial Setup
```bash
kind create cluster
kubectl apply -f deploy/crds/cluster-operator.infobloxopen.github.com_clusters_crd.yaml
kubectl get crds
```
```bash
kubectl create ns kops && kubectl config set-context $(kubectl config current-context) --namespace=kops
```

#### Change & Debug
If changes to types
```bash
operator-sdk generate k8s
```

Build and Run
```bash
make deploy-local
```

Test to see if working
```bash
make cluster
```
If you stop and make changes and rerun controller:
```bash
make operator-todo
```
### Cluster Testing
Assuming you have minikube or cluster with helm tiller you can run
Build and Run
```bash
make deploy
```

Test to see if working
```bash
make cluster
```

#### Debugging
Getting debugging to work with Delve is important, go the latest version
```bash
go get -u github.com/go-delve/delve 
``` 
If you want to run from command line you can build binary:
```bash
cd $GOPATH/src/github.com/go-delve/delve
make install
```
Debugging with a client-go operator was easier since
you can build the project native, but with the sdk-operator you have to build
with delve option
```bash
make operator-debug
```

Then connect with remote debugger (even though you are running it local). The default
port 2345 is what you need for the port, later if you run operator in cluster you need
to forward the port and you will need to configure a higher number port, more on this later.



## Kops
The cluster-operator uses kops for creating clusters on AWS. The base requirements are:
- AWS Key Pair with proper AWS-IAM for Kops
- AWS S3 store for Kops state

In the following examples we will create DNS Route53 and use the default VPC but these and
other AWS services should be managed by cluster-operator using
[AWS Service Broker](https://aws.amazon.com/blogs/opensource/building-own-service-broker-services/) or
[AWS Service Operator](https://github.com/amazon-archives/aws-service-operator)
I have made a decision which to use yet, the later is preferred although AWS has
commited to support 
[AWS Service Operator as a product recently](https://github.com/aws/aws-service-operator-k8s).

### Kops Basics
Here is a simple cluster to create, I am creating them in two AZs:
```bash
aws ec2 describe-availability-zones --region us-east-2 | grep ZoneName
            "ZoneName": "us-east-2a"
            "ZoneName": "us-east-2b"
            "ZoneName": "us-east-2c"
```
AWS settings
```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-east-2
```

Kops settings
```bash
export KOPS_CLUSTER_NAME=seizadicluster.soheil.belamaric.com
export KOPS_STATE_STORE=s3://kops.state.seizadi.infoblox.com
export VPC_ID=vpc-0a75b33895655b46a
export INTERNET_GATEWAY_ID=igw-047d4259cab6b99d2
```

For now need to create the S3 state storage:
```bash
$ aws s3 cb kops.state.seizadi.infoblox.com
```

Create VPC
```bash
aws ec2 create-vpc --cidr-block 172.10.16.0/16 --region ${AWS_DEFAULT_REGION}
```

Create IGW
```bash
aws ec2 create-internet-gateway --region ${AWS_DEFAULT_REGION}
aws ec2 attach-internet-gateway --internet-gateway-id ${INTERNET_GATEWAY_ID} --vpc-id ${VPC_ID} --region ${AWS_DEFAULT_REGION}
```
```bash
kops create cluster \
--name=${KOPS_CLUSTER_NAME} \
--state=${KOPS_STATE_STORE} \
 --ssh-public-key=kops.pub \
--vpc=${VPC_ID} \
--master-count 1 \
--master-size=t2.micro \
--node-count=2 \
--node-size=t2.micro \
--zones=us-east-2a,us-east-2b
```

This creates a desired state, following to create it, or you can do the same
with 'kops create --yes':
```bash
kops update cluster --yes
```
Then to check the status:
```bash
kops validate cluster -o json
```
```json
{
   "failures":[
      {
         "type":"dns",
         "name":"apiserver",
         "message":"Validation Failed\n\nThe dns-controller Kubernetes deployment has not updated the Kubernetes cluster's API DNS entry to the correct IP address.  The API DNS IP address is the placeholder address that kops creates: 203.0.113.123.  Please wait about 5-10 minutes for a master to start, dns-controller to launch, and DNS to propagate.  The protokube container and dns-controller deployment logs may contain more diagnostic information.  Etcd and the API DNS entries must be updated for a kops Kubernetes cluster to start."
      }
   ]
}
```
When things are OK:
```bash
kops validate cluster -o json | jq
```
```json
{
  "nodes": [
    {
      "name": "ip-172-17-17-143.us-east-2.compute.internal",
      "zone": "us-east-2a",
      "role": "master",
      "hostname": "ip-172-17-17-143.us-east-2.compute.internal",
      "status": "True"
    },
    {
      "name": "ip-172-17-18-77.us-east-2.compute.internal",
      "zone": "us-east-2b",
      "role": "node",
      "hostname": "ip-172-17-18-77.us-east-2.compute.internal",
      "status": "True"
    },
    {
      "name": "ip-172-17-17-247.us-east-2.compute.internal",
      "zone": "us-east-2a",
      "role": "node",
      "hostname": "ip-172-17-17-247.us-east-2.compute.internal",
      "status": "True"
    }
  ]
}
```

To delete cluster when you are done:
```bash
kops delete cluster  --yes
```

### Kops Container
I created a kops container to run the commands:
```bash
docker run \
 -e AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE \
 -e AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
 -e KOPS_CLUSTER_NAME=cluster1.soheil.belamaric.com \
 -e KOPS_STATE_STORE=s3://kops.state.seizadi.infoblox.com \
 soheileizadi/kops:v1.0 create cluster \
 --vpc=vpc-0a75b33895655b46a \
 --node-count=2 \
 --master-size=t2.micro \
 --node-size=t2.micro \
 --ssh-key-name=seizadi_aws \
 --zones=us-east-2a,us-east-2b \
 --master-count 1 
```
```bash
docker run \
 -e AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE \
 -e AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
 -e KOPS_CLUSTER_NAME=cluster1.soheil.belamaric.com \
 -e KOPS_STATE_STORE=s3://kops.state.seizadi.infoblox.com \
 soheileizadi/kops:v1.0 validate cluster -o json
```

### Kops Issues

#### Public DNS
You need public DNS address to get any progress since you need to be able to access
the Cluster, you can may be get around this by running Kops in an EC2 instance in the
boundary but makes development difficult. I started with
cluster1.seizadi-kops.local private DNS with Kops option '--dns private'
moved to cluster1.soheil.belamaric.com for public interface.

There is a gossip-based discovery DNS option for the cluster name.
The only requirement to enable this is to have a cluster ending in k8s.local.

#### SSH Keys
The k8s nodes are based on EC2 instances and kops will need SSH keys to setup access
to them, so Kops needs SSH keys. It will normally find them under (~/.ssh/id_rsa.pub), or
you can set them with option --ssh-public-key which points to a specific location for the
key. There is also secret command, in the example below we create a new ssh public 
key called admin.

```bash
  kops create secret sshpublickey admin -i ~/.ssh/id_rsa.pub \
  --name k8s-cluster.example.com --state s3://example.com
```

There is a better option 
[--ssh-key-name that was added recently](https://github.com/kubernetes/kops/pull/6886), 
it allows you to use AWS SSH keys insead so that they are managed outside of the Kops and
more secure. The downside is that it is not a command line parameter only set in the
Cluster Spec so requires yaml file.

You should consider [AWS System Manager](https://aws.amazon.com/systems-manager/), there
will not be any SSH keys or SSH Port open on EC2 so a better security profile.
There is an option to specify [no SSH Key](https://github.com/kubernetes/kops/pull/7096).
