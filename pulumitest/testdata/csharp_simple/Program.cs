using System.Collections.Generic;
using Pulumi;
using Pulumi.Random;

return await Deployment.RunAsync(() =>
{
    var username = new RandomPet("username");

    return new Dictionary<string, object?>
    {
        ["name"] = username.Id
    };
});
