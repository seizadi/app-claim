# Application Claim Docs

## Introduction
The motiviation for this project was to research application claim pattern.
This pattern allows applciations to have a ligth weight claim that is used
create a much more detailed resource. The deployment environment that is targeted
is Kubernetes and the resources that are targeted by these claims are infrastructure
resources like AWS RDS, S3 and other cloud provider services. The claim pattern
shields the applications from knowing anything about the cloud provider concerns
that are typically managed by a differents team that deploys the applciations
and manages cloud infrastructure.

There are examples of this pattern for example 
[crossplane composition](https://github.com/crossplane/crossplane/blob/master/design/one-pager-composition-revisions.md)
aims to support this claim pattern. Another example of this pattern is the
[db-controller](https://infobloxopen.github.io/db-controller) and its
[DatabaseClaim](https://infobloxopen.github.io/db-controller/#databaseclaim-custom-resource). 
The former crossplane claim pattern is very general purpose and more passive while
the db-controller claim is very specialized for database application and db-controller 
takes a more active role in managing the resource and provides database proxy and 
secret key rotation services.

In this project we will research some other patterns, the goal is to find a common
pattern(s) for application deployment.
