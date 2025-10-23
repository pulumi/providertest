using System.Collections.Generic;
using Pulumi;
using MockSdk;

return await Deployment.RunAsync(() =>
{
    var resource = new MockResource("test-resource", "test-value");

    return new Dictionary<string, object?>
    {
        ["resourceValue"] = resource.Value
    };
});
