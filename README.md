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
operator-sdk add api --api-version=cluster-operator.seizadi.github.com/v1alpha1 --kind=Cluster
operator-sdk add controller  --api-version=cluster-operator.seizadi.github.com/v1alpha1 --kind=Cluster
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
[cluster_types.go](https://github.com/seizadi/cluster-operator/blob/master/pkg/apis/clusteroperator/v1alpha1/cluster_types.go)
and the business logic in
[cluster_controller.go](https://github.com/seizadi/cluster-operator/blob/master/pkg/controller/cluster/cluster_controller.go).

Ater making changes run:
```bash
operator-sdk generate k8s
```
regenerate code using code-gen captured in
[]()

You can custom validation using
[kubebuilder tags](https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html)


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
export KOPS_CLUSTER_NAME=cluster1.soheil.belamaric.com
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
--state=${KOPS_STATE_STORE} \
--vpc=${VPC_ID} \
--node-count=2 \
--master-size=t2.micro \
--node-size=t2.micro \
--zones=us-east-2a,us-east-2b \
--name=${KOPS_CLUSTER_NAME} \
--master-count 1 
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
$ kops validate cluster -o json | jq
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

# Kops Issues

You need public DNS address to get any progress since you need to be able to access
the Cluster, you can may be get around this by running Kops in an EC2 instance in the
boundary but makes development difficult. I started with
cluster1.seizadi-kops.local private DNS with Kops option '--dns private'
moved to cluster1.soheil.belamaric.com for public interface.
