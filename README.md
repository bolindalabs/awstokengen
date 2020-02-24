# awstokengen
CLI script to create temporary AWS credentials based on
[EKS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
for tools which can't be easily migrated to a newer AWS SDK.

**Usage**
```bash
# set AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY and AWS_SESSION_TOKEN environment variable,
# based on the IAM Role referenced in the Kubernetes Service Account.
eval $(awstokengen)

./yourTool # now uses environment variables to authenticate against AWS.
```
