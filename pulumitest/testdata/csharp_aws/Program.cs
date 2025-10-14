using System.Collections.Generic;
using Pulumi;
using Pulumi.Aws.S3;

return await Deployment.RunAsync(() =>
{
    var bucket = new Bucket("my-bucket", new BucketArgs
    {
        Tags = new InputMap<string>
        {
            { "Environment", "test" },
            { "Name", "my-bucket" }
        }
    });

    return new Dictionary<string, object?>
    {
        ["bucketName"] = bucket.Id,
        ["bucketArn"] = bucket.Arn
    };
});
