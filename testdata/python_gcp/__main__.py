import pulumi
import pulumi_gcp as gcp

provider = gcp.Provider("provider")
my_bucket = gcp.storage.Bucket("my-bucket", location="US")
pulumi.export("bucketName", my_bucket.url)
